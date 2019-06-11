```
type Mock struct {
	ID primitive.ObjectID
}

//go:generate datarepo -p PageRequest
type Test interface {
	Create(ctx context.Context, user *Mock) error
	Count(ctx context.Context) (int64, error)
	FindByUsername(ctx context.Context, username string) (*Mock, error)
	FindByKeyOfCredentials(ctx context.Context, key string) (*Mock, error)
	FindManyByUserOrderByID(ctx context.Context, user string, page PageRequest) ([]*Mock, error)
	FindOneByUsernameAndLikeIt(ctx context.Context, username string, it string) (*Mock, error)
}

type PageRequest struct {
	Page int64
	Size int64
}
```


// generated
```
type TestRepository struct {
	collection *mongo.Collection
}

func NewTestRepository(collection *mongo.Collection) *TestRepository {
	return &TestRepository{
		collection,
	}
}
func (z *TestRepository) Create(ctx context.Context, user *Mock) error {

	result, err := z.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil

}
func (z *TestRepository) Count(ctx context.Context) (int64, error) {

	return z.collection.CountDocuments(ctx, nil)

}
func (z *TestRepository) FindByUsername(ctx context.Context, username string) (*Mock, error) {

	result := &Mock{}
	err := z.collection.FindOne(ctx, bson.M{
		"Username": username,
	}, nil).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil

}
func (z *TestRepository) FindByKeyOfCredentials(ctx context.Context, key string) (*Mock, error) {

	result := &Mock{}
	err := z.collection.FindOne(ctx, bson.M{
		"Credentials.Key": key,
	}, nil).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil

}
func (z *TestRepository) FindManyByUserOrderByID(ctx context.Context, user string, page PageRequest) ([]*Mock, error) {

	pageSkip := page.Page * page.Size
	cursor, err := z.collection.Find(ctx, bson.M{
		"$query": bson.M{
			"User": user,
		},
		"$orderBy": bson.M{"ID": 1},
	}, &options.FindOptions{Limit: &page.Size, Skip: &pageSkip})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var list []*Mock
	for cursor.Next(ctx) {
		in := &Mock{}
		if err := cursor.Decode(in); err != nil {
			return nil, err
		}
		list = append(list, in)
	}
	return list, err

}
func (z *TestRepository) FindOneByUsernameAndLikeIt(ctx context.Context, username string, it string) (*Mock, error) {

	result := &Mock{}
	err := z.collection.FindOne(ctx, bson.M{
		"$and": bson.A{
			bson.M{
				"Username": username,
			},
			bson.M{
				"It": bson.M{"$regex": "/" + it + "/"},
			},
		},
	}, nil).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil

}
```