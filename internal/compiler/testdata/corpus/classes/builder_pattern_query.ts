// @expect: SELECT id, name, email FROM users WHERE age >= 18 AND active = 1 ORDER BY name ASC LIMIT 10 OFFSET 20;
class QueryBuilder {
    private table: string = "";
    private fields: string[] = [];
    private whereClauses: string[] = [];
    private orderByField: string | null = null;
    private orderDirection: "ASC" | "DESC" = "ASC";
    private limitCount: number | null = null;
    private offsetCount: number | null = null;

    static from(table: string): QueryBuilder {
        const builder = new QueryBuilder();
        builder.table = table;
        return builder;
    }

    select(...fields: string[]): QueryBuilder {
        this.fields = fields;
        return this;
    }

    where(condition: string): QueryBuilder {
        this.whereClauses.push(condition);
        return this;
    }

    orderBy(field: string, direction: "ASC" | "DESC" = "ASC"): QueryBuilder {
        this.orderByField = field;
        this.orderDirection = direction;
        return this;
    }

    limit(count: number): QueryBuilder {
        this.limitCount = count;
        return this;
    }

    offset(count: number): QueryBuilder {
        this.offsetCount = count;
        return this;
    }

    build(): string {
        const selectPart = this.fields.length > 0 ? this.fields.join(", ") : "*";
        let query = "SELECT " + selectPart + " FROM " + this.table;

        if (this.whereClauses.length > 0) {
            query += " WHERE " + this.whereClauses.join(" AND ");
        }

        if (this.orderByField) {
            query += " ORDER BY " + this.orderByField + " " + this.orderDirection;
        }

        if (this.limitCount !== null) {
            query += " LIMIT " + this.limitCount;
        }

        if (this.offsetCount !== null) {
            query += " OFFSET " + this.offsetCount;
        }

        return query + ";";
    }
}

const sql = QueryBuilder.from("users")
    .select("id", "name", "email")
    .where("age >= 18")
    .where("active = 1")
    .orderBy("name", "ASC")
    .limit(10)
    .offset(20)
    .build();

console.log(sql);
