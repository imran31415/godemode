# Real-World E-Commerce Order Processing Benchmark

## Scenario: Complete Order Fulfillment Pipeline

A customer places an order for 3 items (Laptop, Mouse, Keyboard) with a discount code. The system must process the complete order fulfillment workflow.

### Workflow Steps (12 operations across 4 systems)

#### 1. Customer Validation (CRM System)
- **Tool**: `validateCustomer`
- **Input**: Customer ID, email
- **Output**: Customer status, loyalty tier, address
- **Real-world equivalent**: Salesforce, HubSpot API call

#### 2. Inventory Check (Inventory System)
- **Tool**: `checkInventory`
- **Input**: Product IDs, quantities
- **Output**: Availability status, warehouse locations
- **Real-world equivalent**: SAP, NetSuite inventory query

#### 3. Calculate Shipping (Logistics System)
- **Tool**: `calculateShipping`
- **Input**: Destination, weight, dimensions
- **Output**: Shipping cost, estimated delivery
- **Real-world equivalent**: ShipStation, FedEx API

#### 4. Validate Discount Code (Promotion System)
- **Tool**: `validateDiscount`
- **Input**: Discount code, customer tier, cart total
- **Output**: Discount amount, expiration status
- **Real-world equivalent**: Internal promo engine

#### 5. Calculate Tax (Tax System)
- **Tool**: `calculateTax`
- **Input**: Subtotal, shipping address, product categories
- **Output**: Sales tax amount, tax breakdown
- **Real-world equivalent**: Avalara, TaxJar API

#### 6. Process Payment (Payment System)
- **Tool**: `processPayment`
- **Input**: Total amount, payment method, customer ID
- **Output**: Transaction ID, payment status
- **Real-world equivalent**: Stripe, PayPal API

#### 7. Reserve Inventory (Inventory System)
- **Tool**: `reserveInventory`
- **Input**: Order ID, product IDs, quantities
- **Output**: Reservation IDs, expiration time
- **Real-world equivalent**: Inventory management system

#### 8. Create Shipping Label (Logistics System)
- **Tool**: `createShippingLabel`
- **Input**: Order ID, shipping address, weight
- **Output**: Tracking number, label URL
- **Real-world equivalent**: Shippo, EasyPost API

#### 9. Send Confirmation Email (Email System)
- **Tool**: `sendOrderConfirmation`
- **Input**: Order details, customer email
- **Output**: Email ID, delivery status
- **Real-world equivalent**: SendGrid, Mailchimp API

#### 10. Log Transaction (Analytics System)
- **Tool**: `logTransaction`
- **Input**: Order details, payment info, timestamp
- **Output**: Log ID
- **Real-world equivalent**: Segment, Mixpanel event

#### 11. Update Customer Loyalty Points (CRM System)
- **Tool**: `updateLoyaltyPoints`
- **Input**: Customer ID, purchase amount
- **Output**: New points balance
- **Real-world equivalent**: LoyaltyLion, Smile.io API

#### 12. Create Fulfillment Task (Warehouse System)
- **Tool**: `createFulfillmentTask`
- **Input**: Order ID, warehouse ID, items
- **Output**: Task ID, assigned picker
- **Real-world equivalent**: ShipBob, Fulfillment IQ

## Test Data

```json
{
  "customerId": "CUST-12345",
  "email": "john.doe@example.com",
  "items": [
    {"productId": "PROD-001", "name": "Laptop", "quantity": 1, "price": 1299.99},
    {"productId": "PROD-002", "name": "Mouse", "quantity": 1, "price": 29.99},
    {"productId": "PROD-003", "name": "Keyboard", "quantity": 1, "price": 89.99}
  ],
  "discountCode": "SAVE20",
  "shippingAddress": {
    "street": "123 Main St",
    "city": "San Francisco",
    "state": "CA",
    "zip": "94102"
  },
  "paymentMethod": {
    "type": "credit_card",
    "last4": "4242"
  }
}
```

## Expected Results

```json
{
  "orderId": "ORD-2025-001",
  "status": "confirmed",
  "customer": {
    "id": "CUST-12345",
    "tier": "gold",
    "loyaltyPoints": 1562
  },
  "pricing": {
    "subtotal": 1419.97,
    "discount": 283.99,
    "shipping": 15.00,
    "tax": 91.08,
    "total": 1242.06
  },
  "fulfillment": {
    "trackingNumber": "1Z999AA10123456784",
    "estimatedDelivery": "2025-11-18",
    "warehouse": "WH-SF-01"
  },
  "payment": {
    "transactionId": "txn_1234567890",
    "status": "authorized"
  }
}
```

## Success Criteria

1. ✅ All 12 tools execute successfully
2. ✅ Order total calculated correctly
3. ✅ Payment processed
4. ✅ Inventory reserved
5. ✅ Shipping label created
6. ✅ Customer notified
7. ✅ Complete in < 30 seconds (all approaches)

## Performance Metrics to Compare

### Primary Metrics
- **API Calls to Claude**: How many round trips?
- **Total Duration**: End-to-end time including all tool executions
- **Token Usage**: Input + output tokens
- **Cost**: Based on Claude Sonnet 4 pricing

### Secondary Metrics
- **Time to First Tool Call**: How quickly does execution start?
- **Tool Execution Overhead**: Network calls, serialization, etc.
- **Error Handling**: Ability to recover from failures

## Three Approaches

### 1. Code Mode (GoDeMode)
- **Pattern**: Claude generates complete Go program in 1 API call
- **Tool Access**: Direct registry.Call() in generated code
- **Expected**: 1 API call, ~15-20s total, ~3000-4000 tokens
- **Advantage**: Minimal API latency, full visibility, can use loops/conditionals

### 2. Native Tool Calling
- **Pattern**: Claude makes sequential tool_use calls (Anthropic Messages API)
- **Tool Access**: Claude decides which tool to call, we execute, return result
- **Expected**: 13-15 API calls (plan + 12 tools + summary), ~25-35s total, ~8000-12000 tokens
- **Advantage**: Standard Anthropic approach, easy debugging, error recovery

### 3. Native MCP
- **Pattern**: MCP server exposes tools via JSON-RPC, Claude calls sequentially
- **Tool Access**: HTTP requests to MCP server for each tool
- **Expected**: 3-4 API calls (batching tool_use), ~20-30s total, ~6000-8000 tokens
- **Advantage**: Standard MCP protocol, tool provider agnostic

## Why This Is Realistic

1. **Real Complexity**: 12 operations is typical for order processing
2. **Multiple Systems**: 4 different systems (CRM, Inventory, Logistics, Payment)
3. **Dependencies**: Steps must happen in order (validate → check → calculate → pay → fulfill)
4. **Conditional Logic**: Discount validation depends on customer tier
5. **Error Scenarios**: Payment can fail, inventory can be out of stock
6. **Production Scale**: Similar to Shopify, Amazon, Etsy order flows
