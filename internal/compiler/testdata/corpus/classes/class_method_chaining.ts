// @expect: SELECT id, name FROM users WHERE active = 1
class QueryBuilder {
    private query: string = "";

    select(fields: string): this {
        this.query += "SELECT " + fields + " ";
        return this;
    }

    from(table: string): this {
        this.query += "FROM " + table + " ";
        return this;
    }

    where(cond: string): this {
        this.query += "WHERE " + cond;
        return this;
    }

    build(): string {
        return this.query.trim();
    }
}

const q = new QueryBuilder()
    .select("id, name")
    .from("users")
    .where("active = 1")
    .build();

console.log(q);
