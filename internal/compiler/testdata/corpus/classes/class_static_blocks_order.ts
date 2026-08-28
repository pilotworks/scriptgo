// @expect: BaseComponent.field1 -> BaseComponent.staticBlock1 -> BaseComponent.field2 -> BaseComponent.staticBlock2 -> DerivedComponent.childField -> DerivedComponent.staticBlock
// @expect: val1
// @expect: val2
// @expect: child
const staticLog: string[] = [];

class BaseComponent {
    static field1 = (() => {
        staticLog.push("BaseComponent.field1");
        return "val1";
    })();

    static {
        staticLog.push("BaseComponent.staticBlock1");
    }

    static field2 = (() => {
        staticLog.push("BaseComponent.field2");
        return "val2";
    })();

    static {
        staticLog.push("BaseComponent.staticBlock2");
    }
}

class DerivedComponent extends BaseComponent {
    static childField = (() => {
        staticLog.push("DerivedComponent.childField");
        return "child";
    })();

    static {
        staticLog.push("DerivedComponent.staticBlock");
    }
}

console.log(staticLog.join(" -> "));
console.log(BaseComponent.field1);
console.log(BaseComponent.field2);
console.log(DerivedComponent.childField);
