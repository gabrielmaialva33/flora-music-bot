package database

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"main/internal/utils"
)

// docStore é o motor compartilhado de get→cache→update→modify pra documentos
// _id-keyed do Mongo. Unifica o padrão que BotState e ChatSettings repetiam.
// A chave de cache K NÃO é necessariamente o _id (o BotState cacheia em
// "bot_state" mas o _id no Mongo é "global"), por isso idOf mapeia K → _id.
type docStore[K comparable, T any] struct {
	coll        *mongo.Collection
	cache       *utils.Cache[K, *T]
	idOf        func(K) any // mapeia a chave de cache pro valor de _id no Mongo
	makeDefault func(K) *T  // documento default quando não existe no Mongo
	afterLoad   func(*T)    // hook pós-decode (ex.: buildIndexes); pode ser nil
}

// get devolve o doc cacheado, ou busca no Mongo. Em ErrNoDocuments cacheia e
// devolve o default (sem persistir).
func (s *docStore[K, T]) get(key K) (*T, error) {
	if v, ok := s.cache.Get(key); ok {
		return v, nil
	}

	ctx, cancel := ctx()
	defer cancel()

	var doc T
	err := s.coll.FindOne(ctx, bson.M{"_id": s.idOf(key)}).Decode(&doc)

	if errors.Is(err, mongo.ErrNoDocuments) {
		d := s.makeDefault(key)
		s.cache.Set(key, d)
		return d, nil
	}
	if err != nil {
		return nil, fmt.Errorf("docstore get %v: %w", key, err)
	}

	if s.afterLoad != nil {
		s.afterLoad(&doc)
	}
	s.cache.Set(key, &doc)
	return &doc, nil
}

// update persiste o doc via $set (upsert) e atualiza o cache.
func (s *docStore[K, T]) update(key K, doc *T) error {
	ctx, cancel := ctx()
	defer cancel()

	if _, err := s.coll.UpdateOne(
		ctx,
		bson.M{"_id": s.idOf(key)},
		bson.M{"$set": doc},
		upsertOpt,
	); err != nil {
		return fmt.Errorf("docstore update %v: %w", key, err)
	}

	s.cache.Set(key, doc)
	return nil
}

// modify aplica fn ao doc (mutado in-place sobre o ponteiro cacheado) e só
// persiste se fn retornar true. Se o write falhar, invalida o cache pra não
// servir estado divergente do Mongo até o TTL expirar.
func (s *docStore[K, T]) modify(key K, fn func(*T) bool) error {
	doc, err := s.get(key)
	if err != nil {
		return err
	}

	if fn(doc) {
		if err := s.update(key, doc); err != nil {
			s.cache.Delete(key)
			return err
		}
	}

	return nil
}
