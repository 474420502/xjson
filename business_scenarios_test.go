package xjson

import (
	"math"
	"testing"
	"time"

	"github.com/474420502/xjson/internal/core"
	"github.com/stretchr/testify/assert"
)

// TestEcommerceScenario tests a complete e-commerce JSON processing scenario
func TestEcommerceScenario(t *testing.T) {
	// Sample e-commerce JSON data with supported types
	ecommerceJSON := `{
		"store": {
			"name": "TechShop",
			"location": "New York",
			"products": [
				{
					"id": "p001",
					"name": "Laptop Pro",
					"category": "Electronics",
					"price": 1299.99,
					"stock": 15,
					"in_stock": true,
					"tags": ["laptop", "computer", "pro"],
					"reviews": [
						{
							"user": "alice",
							"rating": 5,
							"comment": "Excellent performance!",
							"date": "2024-01-15T10:30:00Z"
						},
						{
							"user": "bob",
							"rating": 4,
							"comment": "Good value for money",
							"date": "2024-01-20T14:15:00Z"
						}
					]
				},
				{
					"id": "p002",
					"name": "Wireless Mouse",
					"category": "Accessories",
					"price": 29.99,
					"stock": 100,
					"in_stock": true,
					"tags": ["mouse", "wireless", "ergonomic"],
					"reviews": [
						{
							"user": "charlie",
							"rating": 3,
							"comment": "Average mouse",
							"date": "2024-01-10T09:45:00Z"
						}
					]
				},
				{
					"id": "p003",
					"name": "Mechanical Keyboard",
					"category": "Accessories",
					"price": 89.99,
					"stock": 0,
					"in_stock": false,
					"tags": ["keyboard", "mechanical", "gaming"],
					"reviews": []
				}
			]
		},
		"customers": [
			{
				"id": "c001",
				"name": "Alice Johnson",
				"email": "alice@example.com",
				"membership": "premium",
				"orders": [
					{
						"id": "o001",
						"date": "2024-01-15T10:00:00Z",
						"items": [
							{
								"product_id": "p001",
								"quantity": 1,
								"price": 1299.99
							}
						],
						"total": 1299.99
					}
				]
			},
			{
				"id": "c002",
				"name": "Bob Smith",
				"email": "bob@example.com",
				"membership": "standard",
				"orders": []
			}
		]
	}`

	// Parse the JSON
	root, err := Parse(ecommerceJSON)
	assert.NoError(t, err)
	assert.True(t, root.IsValid())

	// Register custom functions
	root.Func("expensive", func(n Node) Node {
		return n.Filter(func(product Node) bool {
			price, _ := product.Get("price").RawFloat()
			return price > 100
		})
	})

	root.Func("highlyRated", func(n Node) Node {
		return n.Filter(func(product Node) bool {
			reviews := product.Get("reviews")
			if reviews.Type() != core.ArrayNode || reviews.Len() == 0 {
				return false
			}

			totalRating := int64(0)
			count := int64(0)
			reviews.ForEach(func(_ interface{}, review Node) {
				totalRating += review.Get("rating").Int()
				count++
			})

			average := float64(totalRating) / float64(count)
			return average >= 4.0
		})
	})

	// Business scenario 1: Find all expensive products
	expensiveProducts := root.Query("/store/products[@expensive]")
	assert.True(t, expensiveProducts.IsValid())
	assert.Equal(t, 1, expensiveProducts.Len()) // Only the laptop is expensive (>100)
	assert.Equal(t, "Laptop Pro", expensiveProducts.Index(0).Get("name").String())

	// Business scenario 2: Find all in-stock products
	inStockProducts := root.Query("/store/products").Filter(func(product Node) bool {
		return product.Get("stock").Int() > 0
	})
	assert.True(t, inStockProducts.IsValid())
	assert.Equal(t, 2, inStockProducts.Len()) // Laptop and mouse are in stock

	// Business scenario 3: Find highly rated products
	highlyRatedProducts := root.Query("/store/products[@highlyRated]")
	assert.True(t, highlyRatedProducts.IsValid())
	assert.Equal(t, 1, highlyRatedProducts.Len()) // Only Laptop has high ratings (4.5 avg)

	// Business scenario 4: Get product names as a string array
	productNames := root.Query("/store/products").Map(func(product Node) interface{} {
		return product.Get("name").String()
	})
	assert.True(t, productNames.IsValid())
	names := make([]string, productNames.Len())
	for i := 0; i < productNames.Len(); i++ {
		names[i] = productNames.Index(i).String()
	}
	assert.Contains(t, names, "Laptop Pro")
	assert.Contains(t, names, "Wireless Mouse")
	assert.Contains(t, names, "Mechanical Keyboard")

	// Business scenario 5: Update stock for a product
	products := root.Get("store").Get("products")
	for i := 0; i < products.Len(); i++ {
		product := products.Index(i)
		if product.Get("name").String() == "Wireless Mouse" {
			product.Set("stock", 95) // Sold 5 mice
			assert.Equal(t, int64(95), product.Get("stock").Int())
			break
		}
	}

	// Business scenario 6: Add a new product
	newProduct := map[string]interface{}{
		"id":       "p004",
		"name":     "USB-C Hub",
		"category": "Accessories",
		"price":    39.99,
		"stock":    50,
		"tags":     []interface{}{"usb", "hub", "adapter"},
		"reviews":  []interface{}{},
	}

	// Get fresh reference to products and append
	products = root.Get("store").Get("products")
	t.Logf("Before append - products length: %d", products.Len())
	result := products.Append(newProduct)
	if result.Error() != nil {
		t.Logf("Append error: %v", result.Error())
	}

	// Refresh the products node after append
	products = root.Get("store").Get("products")
	t.Logf("After append - products length: %d", products.Len())
	if products.Error() != nil {
		t.Logf("Products error: %v", products.Error())
	}
	assert.Equal(t, 4, products.Len())

	// Find the new product
	foundNewProduct := false
	for i := 0; i < products.Len(); i++ {
		product := products.Index(i)
		t.Logf("Product %d name: %s", i, product.Get("name").String())
		if product.Get("name").String() == "USB-C Hub" {
			foundNewProduct = true
			break
		}
	}
	assert.True(t, foundNewProduct)

	// Business scenario 7: Find customers with orders
	customersWithOrders := root.Query("/customers").Filter(func(customer Node) bool {
		return customer.Get("orders").Len() > 0
	})
	assert.Equal(t, 1, customersWithOrders.Len())
	assert.Equal(t, "Alice Johnson", customersWithOrders.Index(0).Get("name").String())

	// Business scenario 8: Get customer emails
	customerEmails := root.Query("/customers").Map(func(customer Node) interface{} {
		return customer.Get("email").String()
	})
	assert.Equal(t, 2, customerEmails.Len())

	// Business scenario 9: Calculate total inventory value
	products = root.Get("store").Get("products") // Get fresh reference
	totalValue := 0.0
	products.ForEach(func(_ interface{}, product Node) {
		price, _ := product.Get("price").RawFloat()
		stock := float64(product.Get("stock").Int())
		totalValue += price * stock
	})
	// Expected: 1299.99*15 + 29.99*95 + 89.99*0 + 39.99*50
	expectedValue := 1299.99*15 + 29.99*95 + 89.99*0 + 39.99*50
	// Use tolerance for floating point comparison
	if diff := math.Abs(expectedValue - totalValue); diff > 1e-6 {
		assert.InDelta(t, expectedValue, totalValue, 1e-6, "inventory total mismatch (diff=%v)", diff)
	}

	// Business scenario 10: Error handling
	nonExistent := root.Query("/store/nonexistent/path")
	assert.False(t, nonExistent.IsValid())
}

// TestAPIScenario tests processing API response data
func TestAPIScenario(t *testing.T) {
	apiResponse := `{
		"status": "success",
		"data": {
			"users": [
				{
					"id": 1,
					"username": "john_doe",
					"email": "john@example.com",
					"active": true,
					"roles": ["user", "editor"],
					"profile": {
						"firstName": "John",
						"lastName": "Doe",
						"age": 30,
						"preferences": {
							"theme": "dark",
							"notifications": true
						}
					},
					"lastLogin": "2024-01-20T15:30:00Z"
				},
				{
					"id": 2,
					"username": "jane_smith",
					"email": "jane@example.com",
					"active": false,
					"roles": ["user"],
					"profile": {
						"firstName": "Jane",
						"lastName": "Smith",
						"age": 25,
						"preferences": {
							"theme": "light",
							"notifications": false
						}
					},
					"lastLogin": "2024-01-18T09:15:00Z"
				},
				{
					"id": 3,
					"username": "admin_user",
					"email": "admin@example.com",
					"active": true,
					"roles": ["user", "admin"],
					"profile": {
						"firstName": "Admin",
						"lastName": "User",
						"age": 35,
						"preferences": {
							"theme": "dark",
							"notifications": true
						}
					},
					"lastLogin": "2024-01-21T10:00:00Z"
				}
			]
		},
		"meta": {
			"total": 3,
			"page": 1,
			"limit": 10
		}
	}`

	root, err := Parse(apiResponse)
	assert.NoError(t, err)
	assert.True(t, root.IsValid())

	// Register functions for API processing
	root.Func("active", func(n Node) Node {
		return n.Filter(func(user Node) bool {
			return user.Get("active").Bool()
		})
	})

	root.Func("admins", func(n Node) Node {
		return n.Filter(func(user Node) bool {
			roles := user.Get("roles")
			isAdmin := false
			if roles.Type() == core.ArrayNode {
				roles.ForEach(func(_ interface{}, role Node) {
					if role.String() == "admin" {
						isAdmin = true
					}
				})
			}
			return isAdmin
		})
	})

	root.Func("recentLogin", func(n Node) Node {
		return n.Filter(func(user Node) bool {
			lastLoginStr, _ := user.Get("lastLogin").RawString()
			lastLogin, err := time.Parse(time.RFC3339, lastLoginStr)
			if err != nil {
				return false
			}
			// Users who logged in after Jan 19, 2024
			cutoff, _ := time.Parse(time.RFC3339, "2024-01-19T00:00:00Z")
			return lastLogin.After(cutoff)
		})
	})

	// Scenario 1: Get all active users
	activeUsers := root.Query("/data/users[@active]")
	assert.Equal(t, 2, activeUsers.Len())

	// Scenario 2: Get admin users
	adminUsers := root.Query("/data/users[@admins]")
	assert.Equal(t, 1, adminUsers.Len())
	assert.Equal(t, "admin_user", adminUsers.Index(0).Get("username").String())

	// Scenario 3: Get recently logged in users
	recentUsers := root.Query("/data/users[@recentLogin]")
	assert.Equal(t, 2, recentUsers.Len()) // john_doe and admin_user

	// Scenario 4: Get user full names
	fullNames := root.Query("/data/users").Map(func(user Node) interface{} {
		firstName := user.Get("profile").Get("firstName").String()
		lastName := user.Get("profile").Get("lastName").String()
		return firstName + " " + lastName
	})
	assert.Equal(t, 3, fullNames.Len())
	names := make([]string, fullNames.Len())
	for i := 0; i < fullNames.Len(); i++ {
		names[i] = fullNames.Index(i).String()
	}
	assert.Contains(t, names, "John Doe")
	assert.Contains(t, names, "Jane Smith")
	assert.Contains(t, names, "Admin User")

	// Scenario 5: Count users by theme preference
	darkThemeCount := 0
	lightThemeCount := 0
	root.Query("/data/users").ForEach(func(_ interface{}, user Node) {
		theme := user.Get("profile").Get("preferences").Get("theme").String()
		switch theme {
		case "dark":
			darkThemeCount++
		case "light":
			lightThemeCount++
		}
	})
	assert.Equal(t, 2, darkThemeCount)
	assert.Equal(t, 1, lightThemeCount)

	// Scenario 6: Update user data
	users := root.Get("data").Get("users")
	for i := 0; i < users.Len(); i++ {
		user := users.Index(i)
		if user.Get("username").String() == "jane_smith" {
			user.Set("active", true)
			assert.True(t, user.Get("active").Bool())
			break
		}
	}

	// Scenario 7: Add a new role to a user
	users = root.Get("data").Get("users") // Refresh reference
	for i := 0; i < users.Len(); i++ {
		user := users.Index(i)
		if user.Get("username").String() == "admin_user" {
			// Get current roles and add a new one
			currentRoles := user.Get("roles").Array()
			newRoles := make([]interface{}, len(currentRoles)+1)
			for j, role := range currentRoles {
				newRoles[j] = role.String()
			}
			newRoles[len(newRoles)-1] = "moderator"

			user.Set("roles", newRoles)
			// Refresh reference
			user = users.Index(i)
			assert.Equal(t, 3, user.Get("roles").Len())

			// Verify the new role exists
			hasModerator := false
			user.Get("roles").ForEach(func(_ interface{}, role Node) {
				if role.String() == "moderator" {
					hasModerator = true
				}
			})
			assert.True(t, hasModerator)
			break
		}
	}
}

// TestConfigScenario tests configuration file processing
func TestConfigScenario(t *testing.T) {
	configJSON := `{
		"app": {
			"name": "MyApplication",
			"version": "1.2.3",
			"environment": "production",
			"debug": false
		},
		"server": {
			"host": "0.0.0.0",
			"port": 8080,
			"tls": {
				"enabled": true,
				"certFile": "/etc/ssl/cert.pem",
				"keyFile": "/etc/ssl/key.pem"
			}
		},
		"database": {
			"host": "localhost",
			"port": 5432,
			"name": "myapp",
			"user": "dbuser",
			"password": "dbpass",
			"pool": {
				"maxConnections": 20,
				"minConnections": 5,
				"maxLifetimeMinutes": 60
			}
		},
		"logging": {
			"level": "info",
			"format": "json",
			"outputs": [
				{
					"type": "file",
					"path": "/var/log/myapp.log",
					"rotation": {
						"maxSizeMB": 100,
						"maxAgeDays": 30,
						"compress": true
					}
				},
				{
					"type": "stdout",
					"color": true
				}
			]
		},
		"features": {
			"authentication": {
				"enabled": true,
				"providers": ["local", "oauth2"],
				"oauth2": {
					"google": {
						"clientId": "google-client-id",
						"clientSecret": "google-client-secret",
						"redirectUrl": "https://myapp.com/auth/google/callback"
					},
					"github": {
						"clientId": "github-client-id",
						"clientSecret": "github-client-secret",
						"redirectUrl": "https://myapp.com/auth/github/callback"
					}
				}
			},
			"caching": {
				"enabled": true,
				"provider": "redis",
				"redis": {
					"host": "localhost",
					"port": 6379,
					"password": "",
					"db": 0
				}
			},
			"monitoring": {
				"enabled": true,
				"metrics": {
					"enabled": true,
					"endpoint": "/metrics"
				},
				"tracing": {
					"enabled": false
				}
			}
		}
	}`

	root, err := Parse(configJSON)
	assert.NoError(t, err)
	assert.True(t, root.IsValid())

	// Register functions for config processing
	root.Func("enabledFeatures", func(n Node) Node {
		return n.Filter(func(feature Node) bool {
			return feature.Get("enabled").Bool()
		})
	})

	root.Func("fileOutputs", func(n Node) Node {
		return n.Filter(func(output Node) bool {
			return output.Get("type").String() == "file"
		})
	})

	root.Func("securePort", func(n Node) Node {
		return n.Filter(func(server Node) bool {
			port := server.Get("port").Int()
			return port > 1024 && port < 49151 // Registered ports range
		})
	})

	// Scenario 1: Validate required configuration sections exist
	requiredSections := []string{"app", "server", "database", "logging"}
	for _, section := range requiredSections {
		sectionNode := root.Get(section)
		assert.True(t, sectionNode.IsValid(), "Missing required section: %s", section)
	}

	// Scenario 2: Check if app is in debug mode
	isDebug := root.Query("/app/debug").Bool()
	assert.False(t, isDebug)

	// Scenario 3: Get server configuration
	serverHost := root.Query("/server/host").String()
	serverPort := root.Query("/server/port").Int()
	assert.Equal(t, "0.0.0.0", serverHost)
	assert.Equal(t, int64(8080), serverPort)

	// Scenario 4: Check TLS configuration
	tlsEnabled := root.Query("/server/tls/enabled").Bool()
	assert.True(t, tlsEnabled)

	certFile := root.Query("/server/tls/certFile").String()
	assert.Equal(t, "/etc/ssl/cert.pem", certFile)

	// Scenario 5: Get database connection info
	dbHost := root.Query("/database/host").String()
	dbPort := root.Query("/database/port").Int()
	dbName := root.Query("/database/name").String()
	assert.Equal(t, "localhost", dbHost)
	assert.Equal(t, int64(5432), dbPort)
	assert.Equal(t, "myapp", dbName)

	// Scenario 6: Check database pool settings
	maxConn := root.Query("/database/pool/maxConnections").Int()
	minConn := root.Query("/database/pool/minConnections").Int()
	maxLifetime := root.Query("/database/pool/maxLifetimeMinutes").Int()
	assert.Equal(t, int64(20), maxConn)
	assert.Equal(t, int64(5), minConn)
	assert.Equal(t, int64(60), maxLifetime)

	// Scenario 7: Get logging configuration
	logLevel := root.Query("/logging/level").String()
	logFormat := root.Query("/logging/format").String()
	assert.Equal(t, "info", logLevel)
	assert.Equal(t, "json", logFormat)

	// Scenario 8: Process logging outputs
	fileOutputs := root.Query("/logging/outputs").Filter(func(output Node) bool {
		return output.Get("type").String() == "file"
	})
	assert.Equal(t, 1, fileOutputs.Len())

	stdoutOutputs := root.Query("/logging/outputs").Filter(func(output Node) bool {
		return output.Get("type").String() == "stdout"
	})
	assert.Equal(t, 1, stdoutOutputs.Len())

	// Scenario 9: Check feature flags
	authEnabled := root.Query("/features/authentication/enabled").Bool()
	cacheEnabled := root.Query("/features/caching/enabled").Bool()
	monitoringEnabled := root.Query("/features/monitoring/enabled").Bool()
	assert.True(t, authEnabled)
	assert.True(t, cacheEnabled)
	assert.True(t, monitoringEnabled)

	// Scenario 10: Get OAuth2 providers
	oauthProviders := root.Query("/features/authentication/providers")
	assert.Equal(t, 2, oauthProviders.Len())

	hasLocal := false
	hasOauth2 := false
	oauthProviders.ForEach(func(_ interface{}, provider Node) {
		switch provider.String() {
		case "local":
			hasLocal = true
		case "oauth2":
			hasOauth2 = true
		}
	})
	assert.True(t, hasLocal)
	assert.True(t, hasOauth2)

	// Scenario 11: Get OAuth2 configuration
	googleClientId := root.Query("/features/authentication/oauth2/google/clientId").String()
	githubClientId := root.Query("/features/authentication/oauth2/github/clientId").String()
	assert.Equal(t, "google-client-id", googleClientId)
	assert.Equal(t, "github-client-id", githubClientId)

	// Scenario 12: Update configuration
	// Change log level to debug
	logging := root.Get("logging")
	logging.Set("level", "debug")
	assert.Equal(t, "debug", root.Query("/logging/level").String())

	// Add a new OAuth2 provider
	oauth2 := root.Get("features").Get("authentication").Get("oauth2")
	oauth2.Set("facebook", map[string]interface{}{
		"clientId":     "facebook-client-id",
		"clientSecret": "facebook-client-secret",
		"redirectUrl":  "https://myapp.com/auth/facebook/callback",
	})

	facebookClientId := root.Query("/features/authentication/oauth2/facebook/clientId").String()
	assert.Equal(t, "facebook-client-id", facebookClientId)

	// Scenario 13: Add a new logging output
	newOutput := map[string]interface{}{
		"type": "syslog",
		"host": "localhost",
		"port": 514,
	}

	loggingOutputs := root.Get("logging").Get("outputs")
	appendResult := loggingOutputs.Append(newOutput)
	if appendResult.Error() != nil {
		// preserve original log semantics
		t.Logf("Append logging output error: %v", appendResult.Error())
	}

	// Refresh the logging outputs node
	loggingOutputs = root.Get("logging").Get("outputs")
	t.Logf("Logging outputs length: %d", loggingOutputs.Len())
	if loggingOutputs.Error() != nil {
		t.Logf("Logging outputs error: %v", loggingOutputs.Error())
	}
	assert.Equal(t, 3, loggingOutputs.Len())

	// Find the new syslog output
	foundSyslog := false
	for i := 0; i < loggingOutputs.Len(); i++ {
		output := loggingOutputs.Index(i)
		t.Logf("Output %d type: %s", i, output.Get("type").String())
		if output.Get("type").String() == "syslog" {
			foundSyslog = true
			break
		}
	}
	assert.True(t, foundSyslog)
}
