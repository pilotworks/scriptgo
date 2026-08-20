class Entity {
    id: number;
    name: string;

    constructor(id: number, name: string) {
        this.id = id;
        this.name = name;
    }

    describe(): string {
        return `Entity[${this.id}]: ${this.name}`;
    }

    getPower(): number {
        return 10;
    }
}

class Monster extends Entity {
    level: number;

    constructor(id: number, name: string, level: number) {
        super(id, name);
        this.level = level;
    }

    describe(): string {
        return `Monster[${this.id}]: ${this.name} (Lvl ${this.level})`;
    }

    getPower(): number {
        return super.getPower() * this.level;
    }
}

class BossMonster extends Monster {
    bossRank: number;

    constructor(id: number, name: string, level: number, bossRank: number) {
        super(id, name, level);
        this.bossRank = bossRank;
    }

    describe(): string {
        return `BOSS[Rank ${this.bossRank}] -> ${super.describe()}`;
    }

    getPower(): number {
        return super.getPower() * this.bossRank * 2;
    }
}

const e: Entity = new Entity(1, "NPC");
console.log(e.describe());
console.log(e.getPower());

const m: Monster = new Monster(2, "Goblin", 5);
console.log(m.describe());
console.log(m.getPower());

const b: BossMonster = new BossMonster(3, "Dragon", 50, 3);
console.log(b.describe());
console.log(b.getPower());

function printEntitySummary(entity: Entity): void {
    console.log(`Summary: ${entity.describe()} with power ${entity.getPower()}`);
}

printEntitySummary(e);
printEntitySummary(m);
printEntitySummary(b);
