// @expect: === State Machine & Event Bus Integration Test ===
// @expect: Order 1 initial state: PENDING
// @expect: Order 1 final state: DELIVERED, Tracking: TRACK-VN99281
// @expect: Order 2 final state: CANCELLED, Reason: Item out of stock
// @expect: Invalid transition after cancel succeeded: false
// @expect: 
// @expect: Transition Audit Log:
// @expect: [EVENT] Order ORD-101: PENDING -> PAYMENT_RECEIVED (on PAY)
// @expect: [EVENT] Order ORD-101: PAYMENT_RECEIVED -> PACKING (on PACK)
// @expect: [EVENT] Order ORD-101: PACKING -> SHIPPED (on SHIP)
// @expect: [EVENT] Order ORD-101: SHIPPED -> DELIVERED (on DELIVER)
// @expect: [EVENT] Order ORD-102: PENDING -> PAYMENT_RECEIVED (on PAY)
// @expect: [EVENT] Order ORD-102: PAYMENT_RECEIVED -> CANCELLED (on CANCEL)

// State Machine, Event Bus & Asynchronous Pipeline Processing Scenario

type EventListener<T> = (data: T) => void;

class TypedEventEmitter {
    private events: Map<string, Function[]> = new Map();

    on<T>(event: string, listener: EventListener<T>): void {
        if (!this.events.has(event)) {
            this.events.set(event, []);
        }
        this.events.get(event)!.push(listener);
    }

    emit<T>(event: string, data: T): void {
        const listeners = this.events.get(event);
        if (listeners) {
            for (let i = 0; i < listeners.length; i++) {
                listeners[i](data);
            }
        }
    }
}

// Order Fulfillment Finite State Machine (FSM)
enum OrderState {
    PENDING = "PENDING",
    PAYMENT_RECEIVED = "PAYMENT_RECEIVED",
    PACKING = "PACKING",
    SHIPPED = "SHIPPED",
    DELIVERED = "DELIVERED",
    CANCELLED = "CANCELLED"
}

enum OrderEvent {
    PAY = "PAY",
    PACK = "PACK",
    SHIP = "SHIP",
    DELIVER = "DELIVER",
    CANCEL = "CANCEL"
}

interface OrderContext {
    orderId: string;
    amount: number;
    trackingNumber?: string;
    cancelledReason?: string;
}

class OrderStateMachine {
    private currentState: OrderState = OrderState.PENDING;
    public readonly context: OrderContext;
    private bus: TypedEventEmitter;

    constructor(context: OrderContext, bus: TypedEventEmitter) {
        this.context = context;
        this.bus = bus;
    }

    getState(): OrderState {
        return this.currentState;
    }

    transition(event: OrderEvent, payload?: string): boolean {
        const prev = this.currentState;
        let next: OrderState | null = null;

        switch (this.currentState) {
            case OrderState.PENDING:
                if (event === OrderEvent.PAY) next = OrderState.PAYMENT_RECEIVED;
                if (event === OrderEvent.CANCEL) {
                    next = OrderState.CANCELLED;
                    this.context.cancelledReason = payload || "User requested cancellation";
                }
                break;

            case OrderState.PAYMENT_RECEIVED:
                if (event === OrderEvent.PACK) next = OrderState.PACKING;
                if (event === OrderEvent.CANCEL) {
                    next = OrderState.CANCELLED;
                    this.context.cancelledReason = payload || "Refunded before packing";
                }
                break;

            case OrderState.PACKING:
                if (event === OrderEvent.SHIP) {
                    next = OrderState.SHIPPED;
                    this.context.trackingNumber = payload || "TRACK_DEFAULT";
                }
                break;

            case OrderState.SHIPPED:
                if (event === OrderEvent.DELIVER) next = OrderState.DELIVERED;
                break;

            case OrderState.DELIVERED:
            case OrderState.CANCELLED:
                // Terminal states
                break;
        }

        if (next !== null) {
            this.currentState = next;
            this.bus.emit("order_transition", {
                orderId: this.context.orderId,
                from: prev,
                to: next,
                event
            });
            return true;
        }

        return false;
    }
}

function main(): void {
    console.log("=== State Machine & Event Bus Integration Test ===");

    const bus = new TypedEventEmitter();

    const transitionLog: string[] = [];
    bus.on("order_transition", (data: { orderId: string; from: string; to: string; event: string }) => {
        transitionLog.push(`[EVENT] Order ${data.orderId}: ${data.from} -> ${data.to} (on ${data.event})`);
    });

    // 1. Successful Fulfillment Flow
    const order1 = new OrderStateMachine({ orderId: "ORD-101", amount: 250 }, bus);
    console.log(`Order 1 initial state: ${order1.getState()}`);

    order1.transition(OrderEvent.PAY);
    order1.transition(OrderEvent.PACK);
    order1.transition(OrderEvent.SHIP, "TRACK-VN99281");
    order1.transition(OrderEvent.DELIVER);

    console.log(`Order 1 final state: ${order1.getState()}, Tracking: ${order1.context.trackingNumber}`);

    // 2. Cancellation Flow
    const order2 = new OrderStateMachine({ orderId: "ORD-102", amount: 50 }, bus);
    order2.transition(OrderEvent.PAY);
    order2.transition(OrderEvent.CANCEL, "Item out of stock");

    console.log(`Order 2 final state: ${order2.getState()}, Reason: ${order2.context.cancelledReason}`);

    // 3. Invalid Transition
    const invalidAttempt = order2.transition(OrderEvent.SHIP);
    console.log(`Invalid transition after cancel succeeded: ${invalidAttempt}`);

    console.log("\nTransition Audit Log:");
    for (let i = 0; i < transitionLog.length; i++) {
        console.log(transitionLog[i]);
    }
}

main();
