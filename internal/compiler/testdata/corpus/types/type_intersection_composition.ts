// @expect: ID: 101, Tag: primary, Version: 2, Active: true
// @expect: Total combined: 42

interface Entity {
  id: number;
  version: number;
}

interface Metadata {
  tag: string;
  active: boolean;
}

type EntityRecord = Entity & Metadata;

function processRecord(rec: EntityRecord): string {
  return `ID: ${rec.id}, Tag: ${rec.tag}, Version: ${rec.version}, Active: ${rec.active}`;
}

const item: EntityRecord = {
  id: 101,
  version: 2,
  tag: "primary",
  active: true,
};

console.log(processRecord(item));

type NumA = { a: number };
type NumB = { b: number };
type NumPair = NumA & NumB;

function sumPair(p: NumPair): number {
  return p.a + p.b;
}

const pair: NumPair = { a: 20, b: 22 };
console.log(`Total combined: ${sumPair(pair)}`);
