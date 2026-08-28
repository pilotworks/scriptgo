// @expect: Base[v1] -> Middle[v2] -> Derived[v3]
// @expect: 60

class BaseComponent {
  tag: string;
  constructor(tag: string) {
    this.tag = tag;
  }

  describe(): string {
    return `Base[${this.tag}]`;
  }

  calculate(val: number): number {
    return val * 2;
  }
}

class MiddleComponent extends BaseComponent {
  middleTag: string;
  constructor(tag: string, middleTag: string) {
    super(tag);
    this.middleTag = middleTag;
  }

  describe(): string {
    return `${super.describe()} -> Middle[${this.middleTag}]`;
  }

  calculate(val: number): number {
    return super.calculate(val) + 10;
  }
}

class DerivedComponent extends MiddleComponent {
  derivedTag: string;
  constructor(tag: string, middleTag: string, derivedTag: string) {
    super(tag, middleTag);
    this.derivedTag = derivedTag;
  }

  describe(): string {
    return `${super.describe()} -> Derived[${this.derivedTag}]`;
  }

  calculate(val: number): number {
    return super.calculate(val) + 30;
  }
}

const comp = new DerivedComponent("v1", "v2", "v3");
console.log(comp.describe());
console.log(comp.calculate(10));
