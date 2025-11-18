package models

type GeoPoint struct {
	Type        string    `bson:"type" json:"type"`               // always "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}
type School struct {
	ID         string   `bson:"_id" json:"_id"`
	Name       string   `bson:"name" json:"name"`
	Address    string   `bson:"address" json:"address"`
	LogoUrl    string   `bson:"logo_url,omitempty" json:"logo_url,omitempty"`
	CreatedAt  string   `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt  string   `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	IsUpdating bool     `bson:"is_updating,omitempty" json:"is_updating,omitempty"`
	IsActive   bool     `bson:"is_active,omitempty" json:"is_active,omitempty"`
	DocFiles   []string `bson:"doc_files,omitempty" json:"doc_files,omitempty"`
	Location   GeoPoint `bson:"location,omitempty" json:"location,omitempty"` // GeoJSON format
	IsDeleted  bool     `bson:"is_deleted,omitempty" json:"is_deleted,omitempty"`
}
