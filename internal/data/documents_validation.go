package data

/*validation_rules_start*/

import "github.com/jempe/mpc/internal/validator"

//var (
//	ErrDuplicateDocumentTitleEn = errors.New("duplicate document title (en)")
//	ErrDuplicateDocumentTitleEs = errors.New("duplicate document title (es)")
//	ErrDuplicateDocumentTitleFr = errors.New("duplicate document title (fr)")
//	ErrDuplicateDocumentURLEn   = errors.New("duplicate document URL (en)")
//	ErrDuplicateDocumentURLEs   = errors.New("duplicate document URL (es)")
//	ErrDuplicateDocumentURLFr   = errors.New("duplicate document URL (fr)")
//	ErrDuplicateDocumentFolder  = errors.New("duplicate document folder")
//)


func ValidateDocument(v *validator.Validator, document *Document, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(document.Name != "", "name", "must be provided")
	//v.Check(len(document.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(document.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func documentCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "documents_title_en_key"`:
//		return ErrDuplicateDocumentTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "documents_title_es_key"`:
//		return ErrDuplicateDocumentTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "documents_title_fr_key"`:
//		return ErrDuplicateDocumentTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "documents_url_en_key"`:
//		return ErrDuplicateDocumentURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "documents_url_es_key"`:
//		return ErrDuplicateDocumentURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "documents_url_fr_key"`:
//		return ErrDuplicateDocumentURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "documents_channel_folder_key"`:
//		return ErrDuplicateDocumentFolder
	default:
		return err
	}
}
/*validation_rules_end*/

