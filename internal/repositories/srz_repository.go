package repositories

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/lva100/go-service/internal/models"
	"github.com/lva100/go-service/pkg/logger"
)

type SrzRepository struct {
	Dbpool       *sql.DB
	CustomLogger *logger.Logger
}

func NewSrzRepository(dbpool *sql.DB, logger *logger.Logger) *SrzRepository {
	return &SrzRepository{
		Dbpool:       dbpool,
		CustomLogger: logger,
	}
}

func (r *SrzRepository) GetMo() ([]string, error) {
	query := `
		  select distinct p.lpu
			from VAtt v
			join [srz3_00].dbo.PEOPLE p on v.pid = p.id
			where MPI = 5532052
			and v.lpudt > p.lpudt and v.lpudx is null and p.lpudx is null
			and  v.LPUPROFILE = 1
			order by p.lpu
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var moItems []string

	rows, err := r.Dbpool.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			r.CustomLogger.Error("Timeout from database", err)
		}
		r.CustomLogger.Error("Database query timed out", err)
	}
	defer rows.Close()

	for rows.Next() {
		var moItem string
		err := rows.Scan(
			&moItem,
		)
		if err != nil {
			r.CustomLogger.Error("SQL Query execute wrong", err)
		}
		moItems = append(moItems, moItem)
	}
	if err = rows.Err(); err != nil {
		r.CustomLogger.Error("SQL Query error", err)
	}
	return moItems, nil
}

func (r *SrzRepository) GetReport() ([]models.Otkrep, error) {
	query := `
		  select v.ENP,
				v.LPU LpuCodeNew,
				LPUNAME LpuNameNew,
				v.LPUDT LpuStart,
				DATEADD(day, -1, v.LPUDT) LpuFinish,
				p.lpu LpuCode
			from VAtt v
			join [srz3_00].dbo.PEOPLE p on v.pid = p.id
			where MPI = 5532052
			and v.lpudt > p.lpudt and v.lpudx is null and p.lpudx is null
			and  v.LPUPROFILE = 1
			order by p.lpu
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var otkrepItems []models.Otkrep

	// rows, err := r.Dbpool.QueryContext(ctx, query, sql.Named("dt_start", from), sql.Named("dt_end", to))
	rows, err := r.Dbpool.QueryContext(ctx, query)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			r.CustomLogger.Error("Timeout from database", err)
		}
		r.CustomLogger.Error("Database query timed out", err)
	}
	defer rows.Close()

	for rows.Next() {
		var otkrepItem models.Otkrep
		err := rows.Scan(
			&otkrepItem.ENP,
			&otkrepItem.LpuCodeNew,
			&otkrepItem.LpuNameNew,
			&otkrepItem.LpuStart,
			&otkrepItem.LpuFinish,
			&otkrepItem.LpuCode,
		)
		if err != nil {
			r.CustomLogger.Error("SQL Query execute wrong", err)
		}
		otkrepItems = append(otkrepItems, otkrepItem)
	}
	if err = rows.Err(); err != nil {
		r.CustomLogger.Error("SQL Query error", err)
	}
	return otkrepItems, nil
}

func (r *SrzRepository) CreateRequest() error {
	queryTmp := `INSERT INTO MPI_MSG
					VALUES (NULL,getdate(),getdate(),NULL
									,'<mpi:getViewDataAttachStartRequest xmlns:com="http://ffoms.ru/types/$tmp$/commonTypes" xmlns:mpi="http://ffoms.ru/types/$tmp$/mpiAsyncOperationsSchema"><com:externalRequestId>776B09D6-DB4C-4D8A-99D9-47013F7DBCF1</com:externalRequestId><mpi:criteria xmlns:com="http://ffoms.ru/types/$tmp$/commonTypes" xmlns:mpi="http://ffoms.ru/types/$tmp$/mpiAsyncOperationsSchema"><mpi:fieldNameAttached>smo_okato</mpi:fieldNameAttached><mpi:logicOperation>0</mpi:logicOperation><mpi:value>61000</mpi:value></mpi:criteria><mpi:criteria xmlns:com="http://ffoms.ru/types/$tmp$/commonTypes" xmlns:mpi="http://ffoms.ru/types/$tmp$/mpiAsyncOperationsSchema"><mpi:fieldNameAttached>mo_okato</mpi:fieldNameAttached><mpi:logicOperation>1</mpi:logicOperation><mpi:value>61000</mpi:value></mpi:criteria></mpi:getViewDataAttachStartRequest>'
									,NULL,newid(),'getViewDataAttachStartRequest','mpiAsyncOperation'
									,'4,'+FORMAT(GETDATE(), 'yyyyMMdd')
									,'http://10.255.87.30/api/t-foms/integration/ws/24.1.2/wsdl/mpiAsyncOperationsServiceWs'
									,1
									,NULL,'p_mpi_Response_ViewDataAttach'
									,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL)`
	query := strings.Replace(queryTmp, "$tmp$", "24.1.2", -1)
	_, err := r.Dbpool.ExecContext(
		context.Background(),
		query,
	)
	return err
}

func (r *SrzRepository) InsertFLK(fname, fname_i string, phase_k int, status,
	schet_code, ferr_id string, n_line int, n_col, fc_element, code_err, text_err string) error {
	query := `INSERT INTO Flk_err
           (FNAME,FNAME_I,PHASE_K,STATUS,SCHET_CODE
           ,FERR_ID,N_LINE,N_COL,FC_ELEMENT,CODE_ERR
           ,TEXT_FERR)
     VALUES (@FNAME,@FNAME_I,@PHASE_K,@STATUS,@SCHET_CODE
           ,@FERR_ID,@N_LINE,@N_COL,@FC_ELEMENT,@CODE_ERR
           ,@TEXT_FERR)`
	_, err := r.Dbpool.ExecContext(
		context.Background(),
		query,
		sql.Named("FNAME", fname), sql.Named("FNAME_I", fname_i), sql.Named("PHASE_K", phase_k), sql.Named("STATUS", status),
		sql.Named("SCHET_CODE", schet_code), sql.Named("FERR_ID", ferr_id), sql.Named("N_LINE", n_line), sql.Named("N_COL", n_col),
		sql.Named("FC_ELEMENT", fc_element), sql.Named("CODE_ERR", code_err), sql.Named("TEXT_FERR", text_err),
	)
	return err
}

func (r *SrzRepository) InsertLK(fname, fname_i string, phase_k int, status,
	schet_code, lerr_id, n_zap, lerr_code, oshib_id, nsi_schet, lc_element,
	value, parent_el, parent_el_id, text_lerr string) error {
	query := `INSERT INTO Lk_err
           (FNAME,FNAME_I,PHASE_K,STATUS,SCHET_CODE,LERR_ID,N_ZAP,LERR_CODE,
					 OSHIB_ID,NSI_SCHET,LC_ELEMENT,VALUE,PARENT_EL,PARENT_EL_ID,TEXT_LERR)
     VALUES (@FNAME,@FNAME_I,@PHASE_K,@STATUS,@SCHET_CODE,@LERR_ID,@N_ZAP,@LERR_CODE,
					 @OSHIB_ID,@NSI_SCHET,@LC_ELEMENT,@VALUE,@PARENT_EL,@PARENT_EL_ID,@TEXT_LERR)`
	_, err := r.Dbpool.ExecContext(
		context.Background(),
		query,
		sql.Named("FNAME", fname), sql.Named("FNAME_I", fname_i), sql.Named("PHASE_K", phase_k), sql.Named("STATUS", status),
		sql.Named("SCHET_CODE", schet_code), sql.Named("LERR_ID", lerr_id), sql.Named("N_ZAP", n_zap), sql.Named("LERR_CODE", lerr_code),
		sql.Named("OSHIB_ID", oshib_id), sql.Named("NSI_SCHET", nsi_schet), sql.Named("LC_ELEMENT", lc_element), sql.Named("VALUE", value), sql.Named("PARENT_EL", parent_el), sql.Named("PARENT_EL_ID", parent_el_id), sql.Named("TEXT_LERR", text_lerr),
	)
	return err
}

/*
func (r *ErrRepository) GetReport(from, to string) ([]model.UslReport, error) {
	query := "SELECT 1"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	var repItems []model.UslReport

	rows, err := r.Dbpool.QueryContext(ctx, query, sql.Named("dt_start", from), sql.Named("dt_end", to))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout from database", err)
		}
		log.Println("Database query timed out", err)
	}
	defer rows.Close()

	for rows.Next() {
		var repItem model.UslReport
		err := rows.Scan(
			&repItem.Start,
			&repItem.Code_MO,
			&repItem.OrgName,
			&repItem.Code_Usl,
			&repItem.MC,
			&repItem.MF,
			&repItem.Usl_vol,
			&repItem.Usl_fin,
		)
		if err != nil {
			log.Println("SQL Query execute wrong", err)
		}
		repItems = append(repItems, repItem)
	}
	if err = rows.Err(); err != nil {
		log.Println("SQL Query error", err)
	}
	return repItems, nil
}
*/
