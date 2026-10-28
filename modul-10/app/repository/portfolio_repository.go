package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"modul-10/app/model"
)

var (
	ErrPortfolioNotFound = errors.New("portofolio tidak ditemukan")
	ErrProjectLimit      = errors.New("jumlah proyek sudah mencapai batas maksimal 20")
	ErrProjectNotFound   = errors.New("proyek tidak ditemukan")
)

const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, p model.Portfolio) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, project model.Project) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	StatsTechnologies(ctx context.Context) ([]model.TagCount, error)
}

// STRUCT DOKUMEN: Ini bentuk asli di MongoDB (ID pakai ObjectID)
type portfolioDocument struct {
	ID           bson.ObjectID       `bson:"_id,omitempty"`
	StudentID    int                 `bson:"student_id"`
	Headline     string              `bson:"headline"`
	Bio          string              `bson:"bio"`
	Projects     []model.Project     `bson:"projects"`
	Certificates []model.Certificate `bson:"certificates"`
	CreatedAt    time.Time           `bson:"created_at"`
	UpdatedAt    time.Time           `bson:"updated_at"`
}

// Konversi dari dokumen MongoDB ke Model Go
func (d portfolioDocument) toModel() model.Portfolio {
	projects := d.Projects
	if projects == nil {
		projects = []model.Project{}
	}
	certs := d.Certificates
	if certs == nil {
		certs = []model.Certificate{}
	}
	return model.Portfolio{
		ID:           d.ID.Hex(), // ObjectID diubah jadi String
		StudentID:    d.StudentID,
		Headline:     d.Headline,
		Bio:          d.Bio,
		Projects:     projects,
		Certificates: certs,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

type portfolioMongoRepository struct {
	coll *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "student_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("student_portfolio_unique"),
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrPortfolioNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) Upsert(ctx context.Context, p model.Portfolio) (model.Portfolio, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)

	update := bson.M{
		"$set": bson.M{
			"headline":     p.Headline,
			"bio":          p.Bio,
			"certificates": p.Certificates,
			"updated_at":   now,
		},
		"$setOnInsert": bson.M{
			"student_id": p.StudentID,
			"projects":   []model.Project{}, // Array kosong jika baru dibuat
			"created_at": now,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var updatedDoc portfolioDocument // Gunakan document struct, bukan model
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"student_id": p.StudentID}, update, opts).Decode(&updatedDoc)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("menyimpan portofolio: %w", err)
	}
	return updatedDoc.toModel(), nil
}

func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, project model.Project) (model.Portfolio, error) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	limitField := fmt.Sprintf("projects.%d", MaxProjectsPerPortfolio-1)

	update := bson.M{
		"$push": bson.M{"projects": project},
		"$set":  bson.M{"updated_at": now},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedDoc portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, limitField: bson.M{"$exists": false}},
		update, opts,
	).Decode(&updatedDoc)

	if err == nil {
		return updatedDoc.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	exists, _ := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if exists == 0 {
		return model.Portfolio{}, ErrPortfolioNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) error {
	now := time.Now().UTC().Truncate(time.Millisecond)

	update := bson.M{
		"$pull": bson.M{"projects": bson.M{"project_id": projectID}},
		"$set":  bson.M{"updated_at": now},
	}

	res, err := r.coll.UpdateOne(ctx, bson.M{"student_id": studentID}, update)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.ModifiedCount == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *portfolioMongoRepository) StatsTechnologies(ctx context.Context) ([]model.TagCount, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}, {Key: "_id", Value: 1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("menghitung statistik teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var out []struct {
		Tag   string `bson:"_id"`
		Total int    `bson:"total"`
	}
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("membaca statistik: %w", err)
	}

	result := make([]model.TagCount, 0, len(out))
	for _, t := range out {
		result = append(result, model.TagCount{Tag: t.Tag, Total: t.Total})
	}
	return result, nil
}
