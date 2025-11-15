package tools

import (
	"fmt"
	"time"
)

type ToolInfo struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Function    func(map[string]interface{}) (interface{}, error)
}

type Registry struct {
	tools map[string]*ToolInfo
}

func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]*ToolInfo),
	}

	// Register all 12 e-commerce tools
	r.register(&ToolInfo{
		Name:        "validateCustomer",
		Description: "Validate customer and get their information",
		Parameters: map[string]interface{}{
			"customerId": "string",
			"email":      "string",
		},
		Function: validateCustomer,
	})

	r.register(&ToolInfo{
		Name:        "checkInventory",
		Description: "Check product availability in inventory",
		Parameters: map[string]interface{}{
			"products": "array of {productId, quantity}",
		},
		Function: checkInventory,
	})

	r.register(&ToolInfo{
		Name:        "calculateShipping",
		Description: "Calculate shipping cost and delivery estimate",
		Parameters: map[string]interface{}{
			"destination": "object with city, state, zip",
			"weight":      "number (pounds)",
		},
		Function: calculateShipping,
	})

	r.register(&ToolInfo{
		Name:        "validateDiscount",
		Description: "Validate discount code and calculate discount amount",
		Parameters: map[string]interface{}{
			"code":         "string",
			"customerTier": "string",
			"cartTotal":    "number",
		},
		Function: validateDiscount,
	})

	r.register(&ToolInfo{
		Name:        "calculateTax",
		Description: "Calculate sales tax based on location and items",
		Parameters: map[string]interface{}{
			"subtotal": "number",
			"state":    "string",
		},
		Function: calculateTax,
	})

	r.register(&ToolInfo{
		Name:        "processPayment",
		Description: "Process payment and authorize transaction",
		Parameters: map[string]interface{}{
			"amount":     "number",
			"customerId": "string",
			"method":     "string",
		},
		Function: processPayment,
	})

	r.register(&ToolInfo{
		Name:        "reserveInventory",
		Description: "Reserve inventory for confirmed order",
		Parameters: map[string]interface{}{
			"orderId":  "string",
			"products": "array of {productId, quantity}",
		},
		Function: reserveInventory,
	})

	r.register(&ToolInfo{
		Name:        "createShippingLabel",
		Description: "Generate shipping label and tracking number",
		Parameters: map[string]interface{}{
			"orderId": "string",
			"address": "object with street, city, state, zip",
			"weight":  "number",
		},
		Function: createShippingLabel,
	})

	r.register(&ToolInfo{
		Name:        "sendOrderConfirmation",
		Description: "Send order confirmation email to customer",
		Parameters: map[string]interface{}{
			"email":   "string",
			"orderId": "string",
			"total":   "number",
		},
		Function: sendOrderConfirmation,
	})

	r.register(&ToolInfo{
		Name:        "logTransaction",
		Description: "Log transaction details for analytics",
		Parameters: map[string]interface{}{
			"orderId":       "string",
			"customerId":    "string",
			"amount":        "number",
			"transactionId": "string",
		},
		Function: logTransaction,
	})

	r.register(&ToolInfo{
		Name:        "updateLoyaltyPoints",
		Description: "Update customer loyalty points based on purchase",
		Parameters: map[string]interface{}{
			"customerId": "string",
			"amount":     "number",
		},
		Function: updateLoyaltyPoints,
	})

	r.register(&ToolInfo{
		Name:        "createFulfillmentTask",
		Description: "Create warehouse fulfillment task for order",
		Parameters: map[string]interface{}{
			"orderId":     "string",
			"warehouseId": "string",
			"products":    "array of {productId, quantity}",
		},
		Function: createFulfillmentTask,
	})

	return r
}

func (r *Registry) register(tool *ToolInfo) {
	r.tools[tool.Name] = tool
}

func (r *Registry) Call(name string, args map[string]interface{}) (interface{}, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}
	return tool.Function(args)
}

func (r *Registry) List() []*ToolInfo {
	tools := make([]*ToolInfo, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Tool Implementations

func validateCustomer(args map[string]interface{}) (interface{}, error) {
	customerId, _ := args["customerId"].(string)
	email, _ := args["email"].(string)

	// Simulate CRM lookup
	time.Sleep(50 * time.Millisecond)

	return map[string]interface{}{
		"valid":        true,
		"customerId":   customerId,
		"email":        email,
		"tier":         "gold",
		"loyaltyPoints": 1420,
		"address": map[string]interface{}{
			"verified": true,
		},
	}, nil
}

func checkInventory(args map[string]interface{}) (interface{}, error) {
	products, _ := args["products"].([]interface{})

	// Simulate inventory database query
	time.Sleep(75 * time.Millisecond)

	available := make([]map[string]interface{}, 0)
	for _, p := range products {
		product := p.(map[string]interface{})
		available = append(available, map[string]interface{}{
			"productId":   product["productId"],
			"available":   true,
			"warehouse":   "WH-SF-01",
			"stock":       150,
			"reserved":    product["quantity"],
		})
	}

	return map[string]interface{}{
		"allAvailable": true,
		"products":     available,
	}, nil
}

func calculateShipping(args map[string]interface{}) (interface{}, error) {
	destination, _ := args["destination"].(map[string]interface{})
	weight, _ := args["weight"].(float64)

	// Simulate shipping API call
	time.Sleep(100 * time.Millisecond)

	state := destination["state"].(string)
	cost := 10.0
	if state == "CA" {
		cost = 15.0
	}

	return map[string]interface{}{
		"cost":              cost,
		"estimatedDelivery": time.Now().AddDate(0, 0, 3).Format("2006-01-02"),
		"method":            "Standard Ground",
	}, nil
}

func validateDiscount(args map[string]interface{}) (interface{}, error) {
	code, _ := args["code"].(string)
	tier, _ := args["customerTier"].(string)
	cartTotal, _ := args["cartTotal"].(float64)

	// Simulate promo engine lookup
	time.Sleep(40 * time.Millisecond)

	valid := code == "SAVE20"
	discount := 0.0
	if valid {
		if tier == "gold" {
			discount = cartTotal * 0.20 // 20% for gold members
		} else {
			discount = cartTotal * 0.15 // 15% for others
		}
	}

	return map[string]interface{}{
		"valid":    valid,
		"discount": discount,
		"code":     code,
	}, nil
}

func calculateTax(args map[string]interface{}) (interface{}, error) {
	subtotal, _ := args["subtotal"].(float64)
	state, _ := args["state"].(string)

	// Simulate tax calculation service
	time.Sleep(60 * time.Millisecond)

	taxRate := 0.08 // 8% default
	if state == "CA" {
		taxRate = 0.0925 // 9.25% for California
	}

	tax := subtotal * taxRate

	return map[string]interface{}{
		"tax":     tax,
		"rate":    taxRate,
		"state":   state,
	}, nil
}

func processPayment(args map[string]interface{}) (interface{}, error) {
	amount, _ := args["amount"].(float64)
	customerId, _ := args["customerId"].(string)
	method, _ := args["method"].(string)

	// Simulate payment gateway (Stripe, PayPal)
	time.Sleep(200 * time.Millisecond)

	return map[string]interface{}{
		"success":       true,
		"transactionId": fmt.Sprintf("txn_%d", time.Now().Unix()),
		"amount":        amount,
		"customerId":    customerId,
		"method":        method,
		"status":        "authorized",
	}, nil
}

func reserveInventory(args map[string]interface{}) (interface{}, error) {
	orderId, _ := args["orderId"].(string)
	products, _ := args["products"].([]interface{})

	// Simulate inventory reservation
	time.Sleep(80 * time.Millisecond)

	reservations := make([]map[string]interface{}, 0)
	for _, p := range products {
		product := p.(map[string]interface{})
		reservations = append(reservations, map[string]interface{}{
			"productId":     product["productId"],
			"reservationId": fmt.Sprintf("RES-%s-%v", orderId, product["productId"]),
			"expires":       time.Now().Add(24 * time.Hour).Unix(),
		})
	}

	return map[string]interface{}{
		"success":      true,
		"orderId":      orderId,
		"reservations": reservations,
	}, nil
}

func createShippingLabel(args map[string]interface{}) (interface{}, error) {
	orderId, _ := args["orderId"].(string)
	address, _ := args["address"].(map[string]interface{})

	// Simulate shipping label generation (ShipStation, EasyPost)
	time.Sleep(150 * time.Millisecond)

	return map[string]interface{}{
		"success":        true,
		"trackingNumber": fmt.Sprintf("1Z999AA10%d", time.Now().Unix()%1000000000),
		"labelUrl":       fmt.Sprintf("https://labels.example.com/%s.pdf", orderId),
		"carrier":        "UPS",
	}, nil
}

func sendOrderConfirmation(args map[string]interface{}) (interface{}, error) {
	email, _ := args["email"].(string)
	orderId, _ := args["orderId"].(string)

	// Simulate email service (SendGrid, Mailchimp)
	time.Sleep(100 * time.Millisecond)

	return map[string]interface{}{
		"success": true,
		"emailId": fmt.Sprintf("email_%d", time.Now().Unix()),
		"to":      email,
		"subject": fmt.Sprintf("Order Confirmation - %s", orderId),
		"sent":    true,
	}, nil
}

func logTransaction(args map[string]interface{}) (interface{}, error) {
	orderId, _ := args["orderId"].(string)
	customerId, _ := args["customerId"].(string)
	transactionId, _ := args["transactionId"].(string)

	// Simulate analytics logging (Segment, Mixpanel)
	time.Sleep(30 * time.Millisecond)

	return map[string]interface{}{
		"success":       true,
		"logId":         fmt.Sprintf("log_%d", time.Now().Unix()),
		"orderId":       orderId,
		"customerId":    customerId,
		"transactionId": transactionId,
		"timestamp":     time.Now().Unix(),
	}, nil
}

func updateLoyaltyPoints(args map[string]interface{}) (interface{}, error) {
	customerId, _ := args["customerId"].(string)
	amount, _ := args["amount"].(float64)

	// Simulate loyalty program update
	time.Sleep(50 * time.Millisecond)

	currentPoints := 1420
	earnedPoints := int(amount / 10) // $10 = 1 point
	newPoints := currentPoints + earnedPoints

	return map[string]interface{}{
		"success":      true,
		"customerId":   customerId,
		"pointsEarned": earnedPoints,
		"totalPoints":  newPoints,
	}, nil
}

func createFulfillmentTask(args map[string]interface{}) (interface{}, error) {
	orderId, _ := args["orderId"].(string)
	warehouseId, _ := args["warehouseId"].(string)
	products, _ := args["products"].([]interface{})

	// Simulate warehouse management system
	time.Sleep(70 * time.Millisecond)

	return map[string]interface{}{
		"success":     true,
		"taskId":      fmt.Sprintf("TASK-%s", orderId),
		"warehouseId": warehouseId,
		"assignedTo":  "picker-42",
		"priority":    "normal",
		"itemCount":   len(products),
	}, nil
}
