package data

/*validation_rules_start*/

import "github.com/jempe/mpc/internal/validator"

//var (
//	ErrDuplicateActorTitleEn = errors.New("duplicate actor title (en)")
//	ErrDuplicateActorTitleEs = errors.New("duplicate actor title (es)")
//	ErrDuplicateActorTitleFr = errors.New("duplicate actor title (fr)")
//	ErrDuplicateActorURLEn   = errors.New("duplicate actor URL (en)")
//	ErrDuplicateActorURLEs   = errors.New("duplicate actor URL (es)")
//	ErrDuplicateActorURLFr   = errors.New("duplicate actor URL (fr)")
//	ErrDuplicateActorFolder  = errors.New("duplicate actor folder")
//)


func ValidateActor(v *validator.Validator, actor *Actor, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(actor.Name != "", "name", "must be provided")
	//v.Check(len(actor.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(actor.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func actorCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "actors_title_en_key"`:
//		return ErrDuplicateActorTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "actors_title_es_key"`:
//		return ErrDuplicateActorTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "actors_title_fr_key"`:
//		return ErrDuplicateActorTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "actors_url_en_key"`:
//		return ErrDuplicateActorURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "actors_url_es_key"`:
//		return ErrDuplicateActorURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "actors_url_fr_key"`:
//		return ErrDuplicateActorURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "actors_channel_folder_key"`:
//		return ErrDuplicateActorFolder
	default:
		return err
	}
}
/*validation_rules_end*/

