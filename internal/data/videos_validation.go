package data

/*validation_rules_start*/

import "github.com/jempe/mpc/internal/validator"

//var (
//	ErrDuplicateVideoTitleEn = errors.New("duplicate video title (en)")
//	ErrDuplicateVideoTitleEs = errors.New("duplicate video title (es)")
//	ErrDuplicateVideoTitleFr = errors.New("duplicate video title (fr)")
//	ErrDuplicateVideoURLEn   = errors.New("duplicate video URL (en)")
//	ErrDuplicateVideoURLEs   = errors.New("duplicate video URL (es)")
//	ErrDuplicateVideoURLFr   = errors.New("duplicate video URL (fr)")
//	ErrDuplicateVideoFolder  = errors.New("duplicate video folder")
//)


func ValidateVideo(v *validator.Validator, video *Video, action int) {
	//if action == validator.ActionCreate {
	//	if genericItem.GenericCategoryID == 0 {
	//		genericItem.GenericCategoryID = 1
	//	}
	//}

	//v.Check(genericItem.GenericCategoryID > 0, "generic_category_id", "must be set")
	//v.Check(video.Name != "", "name", "must be provided")
	//v.Check(len(video.Name) >= 3, "name", "must be at least 3 bytes long")
	//v.Check(len(video.Name) <= 200, "name", "must not be more than 200 bytes long")
}

func videoCustomError(err error) error {
	switch {
//	case err.Error() == `pq: duplicate key value violates unique constraint "videos_title_en_key"`:
//		return ErrDuplicateVideoTitleEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "videos_title_es_key"`:
//		return ErrDuplicateVideoTitleEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "videos_title_fr_key"`:
//		return ErrDuplicateVideoTitleFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "videos_url_en_key"`:
//		return ErrDuplicateVideoURLEn
//	case err.Error() == `pq: duplicate key value violates unique constraint "videos_url_es_key"`:
//		return ErrDuplicateVideoURLEs
//	case err.Error() == `pq: duplicate key value violates unique constraint "videos_url_fr_key"`:
//		return ErrDuplicateVideoURLFr
//	case err.Error() == `pq: duplicate key value violates unique constraint "videos_channel_folder_key"`:
//		return ErrDuplicateVideoFolder
	default:
		return err
	}
}
/*validation_rules_end*/

