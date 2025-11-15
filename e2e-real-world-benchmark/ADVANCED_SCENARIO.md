# Advanced Scenario: Fraud Detection & Risk-Based Order Processing

## Overview

This scenario pushes the limits with **25+ operations** involving:
- 🔄 **Conditional logic** - Different paths based on fraud scores
- 🔁 **Loops** - Analyzing transaction history (10+ items)
- ⚠️ **Error handling** - Multiple potential failure points
- 🧮 **Complex calculations** - Multi-factor risk scoring
- 🌳 **Decision trees** - Multi-level approval workflows
- 📊 **Data transformations** - Aggregating, filtering, scoring

## The Scenario

A **$15,000 order** from a new customer triggers advanced fraud detection before processing.

### Order Details
```json
{
  "orderId": "ORD-2025-HIGH-001",
  "customerId": "CUST-NEW-789",
  "email": "suspicious@tempmail.xyz",
  "totalAmount": 15000.00,
  "items": [
    {"productId": "PROD-ELECT-001", "name": "MacBook Pro", "price": 2499.99, "qty": 3},
    {"productId": "PROD-ELECT-002", "name": "iPhone 15 Pro", "price": 1199.99, "qty": 5},
    {"productId": "PROD-ELECT-003", "name": "iPad Pro", "price": 999.99, "qty": 3}
  ],
  "shippingAddress": {
    "street": "123 Temporary St",
    "city": "Miami",
    "state": "FL",
    "zip": "33101",
    "country": "US"
  },
  "billingAddress": {
    "street": "456 Different Ave",
    "city": "Los Angeles",
    "state": "CA",
    "zip": "90001",
    "country": "US"
  },
  "paymentMethod": {
    "type": "credit_card",
    "last4": "1234",
    "issuer": "Unknown Bank"
  },
  "deviceFingerprint": "new_device_unknown_browser",
  "ipAddress": "203.0.113.45",
  "orderTime": "2025-11-15T03:47:23Z"
}
```

### Red Flags Detected
1. ⚠️ **New customer** - Account created 2 hours ago
2. ⚠️ **High value** - $15,000 (10x average order)
3. ⚠️ **High quantity electronics** - Common fraud target
4. ⚠️ **Mismatched addresses** - Billing ≠ Shipping
5. ⚠️ **Suspicious email** - Temp mail service
6. ⚠️ **Odd timing** - 3:47 AM order
7. ⚠️ **New device** - Never seen before
8. ⚠️ **Suspicious IP** - Known proxy/VPN

## Complete Workflow (25+ Operations)

### Phase 1: Initial Validation (4 operations)
1. **validateCustomerAccount** - Check account age, history, verification status
2. **validateEmail** - Check email domain, disposable email detection
3. **validatePhoneNumber** - Verify phone, check carrier, SMS verification
4. **checkDeviceFingerprint** - Analyze device, browser, location consistency

### Phase 2: Fraud Analysis (8 operations)
5. **getCustomerTransactionHistory** - Retrieve all past orders (LOOP: analyze each)
6. **calculateAverageOrderValue** - Compare to current order
7. **checkVelocityRules** - Orders in last 24hrs, 7days, 30days
8. **analyzeIPAddress** - Geolocation, proxy detection, reputation check
9. **checkAddressMismatch** - Billing vs shipping distance, history
10. **verifyPaymentMethod** - Card verification, issuer check, BIN lookup
11. **checkProductRiskScore** - High-value electronics flagged
12. **calculateFraudScore** - Multi-factor risk algorithm (0-100)

### Phase 3: Decision Logic (3 operations)
13. **determineFraudLevel** - Low (<30), Medium (30-70), High (>70)
14. **getRequiredVerifications** - Based on fraud level
15. **checkManualReviewQueue** - If score > 70, queue for human review

### Phase 4: Enhanced Verification (IF fraud score > 30) (5 operations)
16. **sendSMSVerification** - Two-factor authentication
17. **verifySMSCode** - Check user response
18. **requestIDVerification** - Ask for photo ID
19. **checkAgainstWatchlist** - Compare to fraud database
20. **crossReferenceOrders** - Check similar patterns

### Phase 5: Address Verification (4 operations)
21. **validateShippingAddress** - USPS/address verification service
22. **checkAddressHistory** - Previous fraud at this address?
23. **calculateDeliveryRisk** - Freight forwarding, drop shipping detection
24. **suggestSignatureRequired** - Based on value + risk

### Phase 6: Payment Processing (IF approved) (6 operations)
25. **requestPaymentAuthorization** - Pre-auth for full amount
26. **applyRiskBasedHold** - Hold funds 3-7 days based on score
27. **setPaymentWatchdog** - Monitor for chargebacks
28. **reserveInventory** - Lock items
29. **calculateInsurance** - Required for high-value shipment
30. **logFraudAnalysis** - Record all scores and decisions

### Phase 7: Order Processing (IF all checks pass) (6+ operations)
31. **createRiskAdjustedShipping** - Signature, insurance, tracking
32. **notifyWarehouse** - Special handling instructions
33. **scheduleDelayedFulfillment** - Wait 24-48hrs for high-risk
34. **sendCustomerNotification** - Order under review message
35. **createMonitoringTask** - Track for suspicious activity
36. **updateRiskProfile** - Customer's long-term fraud score

## The Challenge: Three Approaches

### Approach 1: Code Mode (GoDeMode)

**Expected Performance:**
- **API Calls**: 1-2 (generate code, maybe 1 retry if complex)
- **Duration**: ~15-25s
- **Tokens**: ~5,000-7,000
- **Cost**: ~$0.08-0.12

**Advantages:**
```go
// Claude can generate code with complex logic:

fraudScore := 0.0

// Loop through transaction history
for _, txn := range transactionHistory {
    if txn.Amount > 1000 {
        fraudScore += 10
    }
    if txn.Disputed {
        fraudScore += 25
    }
}

// Conditional verification
if fraudScore > 70 {
    // Queue for manual review
    result, _ := registry.Call("queueManualReview", ...)
    return "MANUAL_REVIEW_REQUIRED"
} else if fraudScore > 30 {
    // Enhanced verification
    sms, _ := registry.Call("sendSMSVerification", ...)
    code, _ := registry.Call("verifySMSCode", ...)
    if !code.Valid {
        return "VERIFICATION_FAILED"
    }
}

// Continue with order processing...
```

**Why It Excels:**
- Can write loops naturally
- Complex conditional logic
- Early returns on failures
- Efficient data transformations
- Single coherent program

### Approach 2: Native Tool Calling

**Expected Performance:**
- **API Calls**: 10-15 (batching some operations)
- **Duration**: ~45-75s
- **Tokens**: ~20,000-30,000
- **Cost**: ~$0.30-0.50

**Challenges:**
```
API Call 1: Plan + Initial validation (4 tools)
  ↓ Results show high fraud risk
API Call 2: Fraud analysis batch 1 (4 tools)
  ↓ Returns transaction history (10 items)
API Call 3: Need to analyze each transaction... (10 tool calls!)
  ↓ Can't loop - must call Claude to decide
API Call 4: Calculate fraud score from results
  ↓ Score is 75 (high risk)
API Call 5: Get verification requirements
  ↓ Needs SMS verification
API Call 6: Send SMS
  ↓ Wait for user...
API Call 7: Verify code
  ↓ Check if valid
API Call 8: If valid, continue verification
  ↓ More tools...
API Call 9-15: Continue based on results...
```

**Why It Struggles:**
- Can't loop efficiently (each iteration needs API call)
- Complex conditionals require multiple roundtrips
- Token usage explodes with history/context
- Sequential nature adds latency
- Hard to maintain state across calls

### Approach 3: Native MCP

**Expected Performance:**
- **API Calls**: 8-12 (to Claude) + 30-40 (to MCP server)
- **Duration**: ~35-60s
- **Tokens**: ~15,000-25,000
- **Cost**: ~$0.25-0.40
- **MCP Overhead**: ~2-3 seconds (network calls)

**Challenges:**
```
API Call 1: tools/list
API Call 2: Batch validation (4 tools)
  ↓ Each tool = HTTP request to MCP server
API Call 3: Fraud analysis (8 tools)
  ↓ 8 × 100ms MCP overhead = 800ms
API Call 4: Decision logic
  ↓ Needs to analyze results
API Call 5-8: Conditional operations based on score
  ↓ More MCP calls...
API Call 9-12: Final processing
```

**Why It's Middle Ground:**
- Better than Tool Calling (some batching)
- Worse than Code Mode (network overhead)
- MCP protocol overhead per tool
- Still struggles with loops/conditionals

## Expected Results Comparison

### Metrics Prediction

| Metric | Code Mode | Tool Calling | Native MCP | Code Mode Advantage |
|--------|-----------|--------------|------------|---------------------|
| **API Calls** | 1-2 | 10-15 | 8-12 (+ 30-40 MCP) | **83-93% fewer** |
| **Duration** | 15-25s | 45-75s | 35-60s | **50-80% faster** |
| **Tokens** | 5,000-7,000 | 20,000-30,000 | 15,000-25,000 | **65-78% fewer** |
| **Cost** | $0.08-0.12 | $0.30-0.50 | $0.25-0.40 | **60-76% cheaper** |
| **Complexity Handling** | ✅ Excellent | ⚠️ Struggles | ⚠️ Moderate | - |

### Why Code Mode Dominates Even More

**1. Loop Efficiency**
- **Code Mode**: `for _, txn := range history { ... }` - instant
- **Tool Calling**: 10 separate API calls to analyze 10 transactions
- **Native MCP**: 10 API calls + 10 MCP HTTP requests

**2. Conditional Logic**
```go
// Code Mode: Natural branching
if fraudScore > 70 {
    return queueManualReview()
} else if fraudScore > 30 {
    if !verifySMS() {
        return "FAILED"
    }
    if !verifyID() {
        return "FAILED"
    }
}
// All in one program!

// Tool Calling: Multiple roundtrips
// API Call 1: Calculate score
// API Call 2: If score > 70, queue review OR
// API Call 3: If score > 30, send SMS
// API Call 4: Check SMS result
// API Call 5: If valid, verify ID
// API Call 6: Check ID result
// ... many more calls
```

**3. Data Transformation**
```go
// Code Mode: Efficient aggregation
totalHighRisk := 0
for _, txn := range history {
    if txn.Amount > 1000 && txn.Recent {
        totalHighRisk++
    }
}

// Tool Calling: Can't aggregate easily
// Must return all data, have Claude analyze in next call
```

## Business Impact at Scale

### High-Risk Order Processing (100 orders/day needing fraud review)

| Approach | Time/Order | Cost/Order | Daily Cost | Annual Cost |
|----------|------------|------------|------------|-------------|
| **Code Mode** | 20s | $0.10 | $10 | **$3,650** |
| **Tool Calling** | 60s | $0.40 | $40 | $14,600 |
| **Native MCP** | 48s | $0.33 | $33 | $12,045 |

**Annual Savings with Code Mode:**
- vs Tool Calling: **$10,950** (75% reduction)
- vs Native MCP: **$8,395** (70% reduction)

### Additional Benefits

**Code Mode:**
- ✅ Handles 100 orders in **33 minutes** (20s each)
- ✅ Can scale to 200+ orders/hour
- ✅ Consistent performance

**Tool Calling:**
- ⚠️ Handles 100 orders in **100 minutes** (60s each)
- ⚠️ Only 60 orders/hour max
- ⚠️ May hit rate limits

**Native MCP:**
- ⚠️ Handles 100 orders in **80 minutes** (48s each)
- ⚠️ ~75 orders/hour max
- ⚠️ Network overhead accumulates

## The Limits Test

### Scenario Complexity Factors

1. **Number of Operations**: 25-35 (vs 12 in simple scenario)
2. **Loops**: Analyzing 10+ transactions
3. **Conditionals**: 5+ decision points
4. **Data Volume**: Transaction history, fraud databases
5. **Error Handling**: Multiple failure paths
6. **State Management**: Tracking scores, flags, results across operations

### Where Each Approach Hits Limits

**Code Mode:**
- ✅ **Scales well** - Code is compact, logic is efficient
- ⚠️ **May need refinement** - Very complex logic might need 2 API calls
- ✅ **Still dominant** - 1-2 calls vs 10-15

**Tool Calling:**
- ❌ **Struggles significantly** - 10-15+ API calls
- ❌ **Token explosion** - Context grows with each call
- ❌ **Latency adds up** - 7s per call × 15 = 105s minimum
- ❌ **Hard to maintain state** - Each call needs full context

**Native MCP:**
- ⚠️ **Better than Tool Calling** - Some batching helps
- ❌ **Network overhead hurts** - 30-40 HTTP requests
- ⚠️ **Protocol complexity** - JSON-RPC serialization per tool
- ⚠️ **Still struggles with loops** - Can't iterate efficiently

## Conclusion

For this **advanced fraud detection scenario** with 25+ operations, loops, and complex logic:

**Code Mode wins by an even larger margin:**
- **80% faster** than Tool Calling
- **60% faster** than Native MCP
- **75% cheaper** than Tool Calling
- **70% cheaper** than Native MCP

**The complexity gap widens:**
- Simple workflow (12 ops): Code Mode 63% faster
- Complex workflow (25+ ops): Code Mode **80% faster**

**Key Insight:** As complexity increases (loops, conditionals, data transformations), Code Mode's architectural advantages compound, while sequential approaches suffer exponentially.

**Real-World Implication:** For fraud detection, insurance processing, complex approvals, or any multi-step decision workflows - Code Mode isn't just better, it's **essential for production viability**.
