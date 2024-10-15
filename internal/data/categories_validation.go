package data

/*validation_rules_start*/

import "github.com/jempe/mpc/internal/validator"

//var (
//	ErrDuplicateCategoryTitleEn = errors.New("duplicate category title (en)")
//	ErrDuplicateCategoryTitleEs = errors.New("duplicate category title (es)")
//	ErrDuplicateCategoryTitleFr = errors.New("duplicate category title (fr)")
//	ErrDuplicateCategoryURLEn   = errors.New("duplicate category URL (en)")
//	ErrDuplicateCategoryURLEs   = errors.New("duplicate category URL (es)")
//	ErrDuplicateCategoryURLFr   = errors.New("duplicate category URL (fr)")
//	ErrDuplicateCategoryFolder  = errors.New("duplicate category folder")
//)


func ValidateCategory(v *validator.Validator, category *Category, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(category.Name != "", "name", "must be provided")
	//v.Check(len(category.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(category.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func categoryCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "categories_title_en_key"`:
//		return ErrDuplicateCategoryTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "categories_title_es_key"`:
//		return ErrDuplicateCategoryTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "categories_title_fr_key"`:
//		return ErrDuplicateCategoryTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "categories_url_en_key"`:
//		return ErrDuplicateCategoryURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "categories_url_es_key"`:
//		return ErrDuplicateCategoryURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "categories_url_fr_key"`:
//		return ErrDuplicateCategoryURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "categories_channel_folder_key"`:
//		return ErrDuplicateCategoryFolder
	default:
		return err
	}
}
/*validation_rules_end*/

