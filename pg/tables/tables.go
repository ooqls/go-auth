package tables

import (
	"fmt"
	"strings"
)

const ContainerUserMapping string = "container_user_mapping"
const CreateContainerUserMappingStmt string = `
CREATE TABLE IF NOT EXISTS "` + ContainerUserMapping + `" (
	"container_id" INT,
	"user_id" INT,
	PRIMARY KEY("container_id", "user_id"),
	CONSTRAINT fk_user
	  FOREIGN KEY("user_id")
		  REFERENCES ` + Users + `("id") ON DELETE CASCADE,
	CONSTRAINT fk_container
	  FOREIGN KEY("container_id")
		  REFERENCES ` + Containers + `("id") ON DELETE CASCADE
);`

const Containers string = "containers"
const CreateContainersStmt string = `
CREATE TABLE IF NOT EXISTS "` + Containers + `" (
	"id" INT GENERATED ALWAYS AS IDENTITY,
	"name" VARCHAR(128),
	"container_type" INT,
	PRIMARY KEY("id")
);`

const Messages string = "messages"
const CreateMessageStmt string = `
CREATE TABLE IF NOT EXISTS "` + Messages + `" ( 
	"id" INT GENERATED ALWAYS AS IDENTITY,
	"message" VARCHAR(256), 
	"from_id" INT,
	"sent_on" VARCHAR(128),
	"container_id" INT,
	PRIMARY KEY("id"),
	CONSTRAINT fk_from
		FOREIGN KEY("from_id")
			REFERENCES ` + Users + `("id") ON DELETE CASCADE,
	CONSTRAINT fk_container
	  FOREIGN KEY ("container_id")
		  REFERENCES ` + Containers + `("id") ON DELETE CASCADE
);`

const CreateMessageSentOnIndexStmt string = `
CREATE INDEX IF NOT EXISTS "from_sent_on_index" ON "` + Messages + `" ("sent_on", "from_id");`
const CreateMessageContainerIdIndexStmt string = `
CREATE INDEX IF NOT EXISTS "container_id_message" ON "` + Messages + `" ("container_id");`
const CreateMessageFromIdToContainerStmt string = `
CREATE INDEX IF NOT EXISTS "from_id_to_container_index" ON "` + Messages + `" ("container_id", "from_id");`
const CreateUserCreatedIndexStmt = `
CREATE INDEX IF NOT EXISTS "user_created_index" ON "` + Users + `" ("created", "id");`

const Users string = "users"
const CreateUsersStmt string = `
CREATE TABLE IF NOT EXISTS "` + Users + `" (
	"id" INT GENERATED ALWAYS AS IDENTITY,
	"name" VARCHAR (128),
	"email" VARCHAR (128),
	"created" DATE,
	PRIMARY KEY("id")
);`

const Passwords string = "passwords"
const CreatePasswordStmt string = `
CREATE TABLE IF NOT EXISTS "` + Passwords + `" (
	"userid" INT,
	"hash_algo" VARCHAR(16),
	"pw_digest" VARCHAR(1024),
	"created" DATE,
	PRIMARY KEY("userid"),
	CONSTRAINT fk_user
	  FOREIGN KEY("userid")
		  REFERENCES ` + Users + `("id") ON DELETE CASCADE
);`

func GetCreateTableStmts() []string {
	return []string{
		CreateUsersStmt,
		CreateContainersStmt,
		CreateMessageStmt,
		CreatePasswordStmt,
		CreateContainerUserMappingStmt,
	}
}

func GetCreateIndexStmts() []string {
	return []string{
		CreateMessageFromIdToContainerStmt,
		CreateMessageContainerIdIndexStmt,
		CreateMessageSentOnIndexStmt,
		CreateUserCreatedIndexStmt,
	}
}

func GetDropTableStmt() string {
	tables := []string{ContainerUserMapping, Passwords, Messages, Users, Containers}
	dropstmt := fmt.Sprintf("DROP TABLE %s", strings.Join(tables, ", "))
	return dropstmt
}
