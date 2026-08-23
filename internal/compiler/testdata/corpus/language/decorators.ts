// ScriptGo Corpus: Static Tier Unified Decorators & Reflection Test Suite
// Consolidated test suite verifying Stage 3 and Legacy experimental decorators in Static Tier.

// --- Context Case: stage3_and_legacy_method_decorators ---
function logged(name: string) {
    console.log("Decorating method " + name);
    return function(target?: (id: number) => string, context?: ClassMethodDecoratorContext) {};
}

class ProductService {
    @logged("fetchProduct")
    fetchProduct(id: number): string {
        return "Product_" + id;
    }
}

// @expect: Decorating method fetchProduct
let svc = new ProductService();
// @expect: Product_42
console.log(svc.fetchProduct(42));

// --- Context Case: reflect_metadata_reflection ---
// @api: Reflect.metadata
// @api: Reflect.getMetadata
// @api: Reflect.hasMetadata
class UserService {
    @Reflect.metadata("role", "admin")
    getUser(id: string): string {
        return "User_" + id;
    }
}

// @expect: true
console.log(Reflect.hasMetadata("role", UserService, "getUser"));

// @expect: admin
console.log(Reflect.getMetadata("role", UserService, "getUser"));

// @expect: function
console.log(Reflect.getMetadata("design:type", UserService, "getUser"));

// @expect: string
console.log(Reflect.getMetadata("design:returntype", UserService, "getUser"));

// --- Context Case: custom_define_metadata ---
// @api: Reflect.defineMetadata
Reflect.defineMetadata("tag", "v1.0", UserService, "getUser");
// @expect: true
console.log(Reflect.hasMetadata("tag", UserService, "getUser"));
// @expect: v1.0
console.log(Reflect.getMetadata("tag", UserService, "getUser"));

// --- Context Case: field_and_class_decorators ---
function trace(tag: string) {
    console.log("Trace decorator applied: " + tag);
    return function(target?: object | Function | undefined, context?: ClassMemberDecoratorContext | ClassDecoratorContext) {};
}

@trace("Order")
class Order {
    @trace("orderId")
    orderId: string = "ORD-001";
}

// @expect: Trace decorator applied: orderId
// @expect: Trace decorator applied: Order
let order = new Order();
// @expect: ORD-001
console.log(order.orderId);


