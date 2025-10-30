package repositories

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lva100/go-service/internal/models"
	"github.com/lva100/go-service/pkg/logger"
)

/*
const queryStr = `
	select h.pid,p.ENP,p.LPU,l.CAPTION,p.LPUDT,h.lpu,h.lpudx
	from [srz3_00].dbo.histlpu h
	join (
			select b.people, max(isnull(h.lpudx,'19000101')) as maxLPUDX
			from [srz3_00].dbo.MO_BUFFER b
			join [srz3_00].dbo.histlpu h on h.pid = b.PEOPLE
			where b.MO_LOG in (
					select ID from [srz3_00].[dbo].[MO_LOG] where DT between '20251027' and DATEADD(day,1,'20251027')
			) and b.people is not null
			group by b.people
	) t on h.PID=t.PEOPLE and h.LPUDX=t.maxLPUDX
	join [srz3_00].dbo.people p on p.ID=h.PID
	join [srz3_00].[dbo].[LPU] l on p.LPU=l.REGNUM
	where h.LPU<>998
	order by pid
`
*/

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

func (r *SrzRepository) GetMo(id int64, dt string) ([]string, error) {
	/*query := `
		  select distinct p.lpu
			from VAtt v
			join [srz3_00].dbo.PEOPLE p on v.pid = p.id
			where MPI = @id
			and v.lpudt > p.lpudt and v.lpudx is null and p.lpudx is null
			and  v.LPUPROFILE = 1
			order by p.lpu
	`*/
	query := `select distinct lpu from (
					select h.lpu
					from [srz3_00].dbo.histlpu h
					join (
							select b.people, max(isnull(h.lpudx,'19000101')) as maxLPUDX
							from [srz3_00].dbo.MO_BUFFER b  
							join [srz3_00].dbo.histlpu h on h.pid = b.PEOPLE
							where b.MO_LOG in (
									select ID from [srz3_00].[dbo].[MO_LOG] 
									where DT between @dt 
													and DATEADD(day,1,@dt)
							) and b.people is not null
							group by b.people
					) t on h.PID=t.PEOPLE and h.LPUDX=t.maxLPUDX
					join [srz3_00].dbo.people p on p.ID=h.PID
					join [srz3_00].[dbo].[LPU] l on p.LPU=l.REGNUM
					where h.LPU<>998
					union all
					select p.lpu
					from VAtt v
					join [srz3_00].dbo.PEOPLE p on v.pid = p.id
					where MPI = @id
					and v.lpudt > p.lpudt and v.lpudx is null and p.lpudx is null
					and  v.LPUPROFILE = 1
					) a
					order by lpu`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var moItems []string

	rows, err := r.Dbpool.QueryContext(ctx, query, sql.Named("id", id), sql.Named("dt", dt))
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

func (r *SrzRepository) GetReport(id int64, dt string) ([]models.Otkrep, error) {
	query := `select h.pid PID,p.ENP,p.LPU LpuCodeNew,l.CAPTION LpuNameNew
									,p.LPUDT LpuStart ,h.lpudx LpuFinish,h.lpu LpuCode
						from [srz3_00].dbo.histlpu h
						join (
								select b.people, max(isnull(h.lpudx,'19000101')) as maxLPUDX
								from [srz3_00].dbo.MO_BUFFER b  
								join [srz3_00].dbo.histlpu h on h.pid = b.PEOPLE
								where b.MO_LOG in (
										select ID from [srz3_00].[dbo].[MO_LOG] 
										where DT between @dt 
														and DATEADD(day,1,@dt)
								) and b.people is not null
								group by b.people
						) t on h.PID=t.PEOPLE and h.LPUDX=t.maxLPUDX
						join [srz3_00].dbo.people p on p.ID=h.PID
						join [srz3_00].[dbo].[LPU] l on p.LPU=l.REGNUM
						where h.LPU<>998
						union all
						select v.pid PID,v.ENP,
							v.LPU LpuCodeNew,
							LPUNAME LpuNameNew,
							v.LPUDT LpuStart,
							DATEADD(day, -1, v.LPUDT) LpuFinish,
							p.lpu LpuCode
						from VAtt v
						join [srz3_00].dbo.PEOPLE p on v.pid = p.id
						where MPI = @id
						and v.lpudt > p.lpudt and v.lpudx is null and p.lpudx is null
						and  v.LPUPROFILE = 1
						order by p.lpu`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var otkrepItems []models.Otkrep

	// rows, err := r.Dbpool.QueryContext(ctx, query, sql.Named("dt_start", from), sql.Named("dt_end", to))
	rows, err := r.Dbpool.QueryContext(ctx, query, sql.Named("id", id), sql.Named("dt", dt))
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
			&otkrepItem.PID,
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

func (r *SrzRepository) CheckRequest(dt string) (int64, bool) {
	query := `select ID
							from MPI_MSG
							where RTYPE='getViewDataAttachStartRequest'
										and SUBSTRING(CARGO,1,1)=4
										and DT between DATEADD(day,-1,@dt) and @dt`
	var id int64
	if err := r.Dbpool.QueryRowContext(
		context.Background(),
		query,
		sql.Named("dt", dt),
	).Scan(&id); err != nil {
		r.CustomLogger.Error("SQL Query error", err)
		return 0, false
	}
	return id, true
}

func (r *SrzRepository) CreateRequest(apiVer string) (int64, error) {
	id := uuid.New()
	queryTmp := `INSERT INTO MPI_MSG
					VALUES (NULL,getdate(),getdate(),NULL
								,'<mpi:getViewDataAttachStartRequest xmlns:com="http://ffoms.ru/types/$tmp$/commonTypes" xmlns:mpi="http://ffoms.ru/types/$tmp$/mpiAsyncOperationsSchema"><com:externalRequestId>$id$</com:externalRequestId><mpi:criteria xmlns:com="http://ffoms.ru/types/$tmp$/commonTypes" xmlns:mpi="http://ffoms.ru/types/$tmp$/mpiAsyncOperationsSchema"><mpi:fieldNameAttached>smo_okato</mpi:fieldNameAttached><mpi:logicOperation>0</mpi:logicOperation><mpi:value>61000</mpi:value></mpi:criteria><mpi:criteria xmlns:com="http://ffoms.ru/types/$tmp$/commonTypes" xmlns:mpi="http://ffoms.ru/types/$tmp$/mpiAsyncOperationsSchema"><mpi:fieldNameAttached>mo_okato</mpi:fieldNameAttached><mpi:logicOperation>1</mpi:logicOperation><mpi:value>61000</mpi:value></mpi:criteria></mpi:getViewDataAttachStartRequest>'
								,NULL,'$id$','getViewDataAttachStartRequest','mpiAsyncOperation'
								,'4,'+FORMAT(GETDATE(), 'yyyyMMdd')
								,'http://10.255.87.30/api/t-foms/integration/ws/$tmp$/wsdl/mpiAsyncOperationsServiceWs'
								,1
								,NULL,'p_mpi_Response_ViewDataAttach'
								,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);
								select ID = convert(bigint, SCOPE_IDENTITY())
								`
	queryTmp = strings.ReplaceAll(queryTmp, "$tmp$", apiVer)
	query := strings.ReplaceAll(queryTmp, "$id$", id.String())
	var lastInseredId int64
	if err := r.Dbpool.QueryRowContext(
		context.Background(),
		query,
	).Scan(&lastInseredId); err != nil {
		r.CustomLogger.Error("SQL Query error", err)
		return 0, err
	}
	return lastInseredId, nil
}

func (r *SrzRepository) InsertLog(pid int, enp, lpuPrikrepCode, lpuPrikrepName string,
	lpudt string, lpuOtkrepCode string, lpudx string, npackage int64, processdt string, filename string) error {
	query := `INSERT INTO [Lpu_History].[dbo].[OtkrepLog]
									(PID,ENP,LPU_PRIK_CODE,LPU_PRIK_NAME,LPUDT
									,LPU_OTKREP_CODE,LPUDX,NPACKAGE,PROCESSDATE,FILENAME)
						VALUES (@PID,@ENP,@LPU_PRIK_CODE,@LPU_PRIK_NAME,@LPUDT
									,@LPU_OTKREP_CODE,@LPUDX,@NPACKAGE,@PROCESSDATE,@FILENAME)`
	_, err := r.Dbpool.ExecContext(
		context.Background(),
		query,
		sql.Named("PID", pid), sql.Named("ENP", enp), sql.Named("LPU_PRIK_CODE", lpuPrikrepCode), sql.Named("LPU_PRIK_NAME", lpuPrikrepName), sql.Named("LPUDT", lpudt), sql.Named("LPU_OTKREP_CODE", lpuOtkrepCode), sql.Named("LPUDX", lpudx), sql.Named("NPACKAGE", npackage), sql.Named("PROCESSDATE", processdt), sql.Named("FILENAME", filename),
	)
	return err
}

func (r *SrzRepository) ClosePrikrep(id int64) error {
	query := `update h
						set 
						h.lpudx = DATEADD(day, -1, v.LPUDT),
						h.REMARK = 'Прикрепление закрыто по письму ФФОМС о двойных прикреплениях'
						from [srz3_00].dbo.histlpu h 
						join VAtt v on v.pid = h.pid
							join [srz3_00].dbo.PEOPLE p on v.pid = p.id
												where MPI = @id
												and v.lpudt > p.lpudt and v.lpudx is null and p.lpudx is null
												and  v.LPUPROFILE = 1
																and p.lpu = h.lpu and h.lpudx is null
						update p
						set 
						p.lpudx = DATEADD(day, -1, v.LPUDT)
						from [srz3_00].dbo.PEOPLE p
						join VAtt v on v.pid = p.id
								where MPI = @id
												and v.lpudt > p.lpudt and v.lpudx is null and p.lpudx is null
												and  v.LPUPROFILE = 1
						`
	_, err := r.Dbpool.ExecContext(
		context.Background(),
		query,
		sql.Named("id", id),
	)
	return err
}
