# Maya
My personal and lightweight go-mysql handler

## Why Maya?
Everybody wants to use an ORM easy to use but sometimes do not satisfy all our whims.  
In order to removing all those unnecessary features and keep a minimal functionality we started to build own one just for fun.  

Also is defined on traditional layered web-api and uses of go-context.

This package is inspired on EF, GORM and others current ORMs.

---------------------------------------
- [DB Context](#db-context)
- [Migrations](#migrations)
- [Entities](#entities)
- [Auditable fields](#auditable-fields)
- [Abstraction](#abstrations)
- [Environment variables](#environment-variables)
- [Support packages](#support-packages)
---------------------------------------

## DB Context
Here is the entry point where you will set connection string, logger and migrations.  
As you can see we are comfortable using _zap_ package. Highly recommended.  
``` go
NewDbContext(server string, dbName string, port string, user string, pass string, logger *zap.Logger, migrations []Migration) *dbContext
```

## Migrations
Migrations are available in Maya just need to define an ID (e.g. 000000001) and array with all your data table definitions.  
Finally pass it to DB Context constructor.
``` go
type Migration struct {
	ID   string   // ID migration. e.g. 00000001
	Up   []string // slice of DDL for creation
	Down []string // slice of DDL for drop
}
```
## Entities
In maya you can define entities using tags (__tname__, __tcol__, __pk__)  
Also, you can set auditable field like :createdBy, createdAt, updatedBy & updatedAt.
``` go
import "github.com/maya"

type User struct {
	maya.IEntity `tname:"users"`                    // db table name
	maya.IAuditable                                 // define this entity is using auditable fields
	ID               int64  `tcol:"id" pk:"true"`   // column name, set as primary key (can be any)
	Email            string `tcol:"email"`
	Fullname         string `tcol:"fullname"`
	PasswordHash     string `tcol:"password_hash"`
	EmailConfirmed   bool   `tcol:"email_confirmed"`
	PhoneConfirmed   bool   `tcol:"phone_confirmed"`
	SecurityStamp    string `tcol:"security_stamp"`
	ConcurrencyStamp string `tcol:"concurrency_stamp"`
	IsLocked         bool   `tcol:"is_locked"`
	IsActive         bool   `tcol:"is_active"`
	Intents          int    `tcol:"intents"`
}
```

## Enqueue/Dequeue service
This service allow you to create locks (DB-based) for entities(table name). The constraint is not invoke any of this methods in middle of DB transaction.  
__Enqueue__ requires an `id` and `entity-pointer`. Optional, you can set extra `info` to add detail.  
To __Dequeue__ just invoke it and all those enqueued items will start to dropping from table.  
We recommend use Enqueue for mutations.
``` go
Enqueue(ctx context.Context, id string, entityPtr any, info string) error
Dequeue(ctx context.Context) error
```

## Auditable fields
Auditable fields are available and just include four field into tables as seen below to save it.  
Also is needed set auditable interface (maya.IAuditable) to detect an entity with this property.
``` sql
`created_by` varchar(100) NOT NULL,
`created_at` datetime NOT NULL,
`update_by` varchar(100) NOT NULL,
`update_at` datetime NOT NULL,
```

## Abstractions
We include 5 basic generic interfaces to start up to sketch db ports/adapters but feel free to implement own ones.
- ICreateRepository
- IUpdateRepository
- IReadAllRepository
- IReadByExternalIDRepository
- IReadSingleRepository

To determine _created_by_ and _updated_by_ is necessary set a value into `context.Context` with key _maya-user-editor_  
Here and example pass it via middleware
``` go
func UseUserInfoMiddl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := "email@mail.com" + ":" + r.RemoteAddr
		ctx := context.WithValue(r.Context(), "maya-user-editor", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```



## Environment variables
|Variable|Description|
|--------|-----------|
|```DB_MIGRATE = false```|Deactive migrations. None or default = active|


## Support packages
- github.com/go-sql-driver/mysql
- github.com/jmoiron/sqlx
- go.uber.org/zap
