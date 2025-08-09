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
	root.RegisterFunc("expensive", func(n Node) Node {
		return n.Filter(func(product Node) bool {
			price, _ := product.Get("price").RawFloat()
			return price > 100
		})
	})

	root.RegisterFunc("highlyRated", func(n Node) Node {
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
	if assert.NoError(t, expensiveProducts.Error()) {
		assert.True(t, expensiveProducts.IsValid())
		assert.Equal(t, 1, expensiveProducts.Len()) // Only the laptop is expensive (>100)
		nameNode := expensiveProducts.Index(0).Get("name")
		if assert.NoError(t, nameNode.Error()) {
			assert.Equal(t, "Laptop Pro", nameNode.String())
		}
	}

	// Business scenario 2: Find all in-stock products
	inStockProducts := root.Query("/store/products")
	if assert.NoError(t, inStockProducts.Error()) {
		filtered := inStockProducts.Filter(func(product Node) bool {
			stock := product.Get("stock")
			return stock.Error() == nil && stock.Int() > 0
		})
		if assert.NoError(t, filtered.Error()) {
			assert.True(t, filtered.IsValid())
			assert.Equal(t, 2, filtered.Len()) // Laptop and mouse are in stock
		}
	}

	// Business scenario 3: Find highly rated products
	highlyRatedProducts := root.Query("/store/products[@highlyRated]")
	if assert.NoError(t, highlyRatedProducts.Error()) {
		assert.True(t, highlyRatedProducts.IsValid())
		assert.Equal(t, 1, highlyRatedProducts.Len()) // Only Laptop has high ratings (4.5 avg)
	}

	// Business scenario 4: Get product names as a string array
	productNamesNode := root.Query("/store/products")
	if assert.NoError(t, productNamesNode.Error()) {
		productNames := productNamesNode.Map(func(product Node) interface{} {
			name := product.Get("name")
			if name.Error() != nil {
				return name.Error()
			}
			return name.String()
		})
		if assert.NoError(t, productNames.Error()) {
			assert.True(t, productNames.IsValid())
			names := make([]string, productNames.Len())
			for i := 0; i < productNames.Len(); i++ {
				nameNode := productNames.Index(i)
				if assert.NoError(t, nameNode.Error()) {
					names[i] = nameNode.String()
				}
			}
			assert.Contains(t, names, "Laptop Pro")
			assert.Contains(t, names, "Wireless Mouse")
			assert.Contains(t, names, "Mechanical Keyboard")
		}
	}

	// Business scenario 5: Update stock for a product
	products := root.Get("store").Get("products")
	if assert.NoError(t, products.Error()) {
		for i := 0; i < products.Len(); i++ {
			product := products.Index(i)
			if !assert.NoError(t, product.Error()) {
				continue
			}

			nameNode := product.Get("name")
			if !assert.NoError(t, nameNode.Error()) {
				continue
			}

			if nameNode.String() == "Wireless Mouse" {
				setNode := product.Set("stock", 95) // Sold 5 mice
				if assert.NoError(t, setNode.Error()) {
					stockNode := product.Get("stock")
					if assert.NoError(t, stockNode.Error()) {
						assert.Equal(t, int64(95), stockNode.Int())
					}
				}
				break
			}
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
	if assert.NoError(t, products.Error()) {
		t.Logf("Before append - products length: %d", products.Len())
		result := products.Append(newProduct)
		assert.NoError(t, result.Error())
	}

	// Refresh the products node after append
	products = root.Get("store").Get("products")
	if assert.NoError(t, products.Error()) {
		t.Logf("After append - products length: %d", products.Len())
		assert.Equal(t, 4, products.Len())
	}

	// Find the new product
	foundNewProduct := false
	if assert.NoError(t, products.Error()) {
		for i := 0; i < products.Len(); i++ {
			product := products.Index(i)
			if !assert.NoError(t, product.Error()) {
				continue
			}
			nameNode := product.Get("name")
			if !assert.NoError(t, nameNode.Error()) {
				continue
			}
			t.Logf("Product %d name: %s", i, nameNode.String())
			if nameNode.String() == "USB-C Hub" {
				foundNewProduct = true
				break
			}
		}
	}
	assert.True(t, foundNewProduct)

	// Business scenario 7: Find customers with orders
	customersWithOrders := root.Query("/customers")
	if assert.NoError(t, customersWithOrders.Error()) {
		filtered := customersWithOrders.Filter(func(customer Node) bool {
			orders := customer.Get("orders")
			return orders.Error() == nil && orders.Len() > 0
		})
		if assert.NoError(t, filtered.Error()) {
			assert.Equal(t, 1, filtered.Len())
			nameNode := filtered.Index(0).Get("name")
			if assert.NoError(t, nameNode.Error()) {
				assert.Equal(t, "Alice Johnson", nameNode.String())
			}
		}
	}

	// Business scenario 8: Get customer emails
	customerEmailsNode := root.Query("/customers")
	if assert.NoError(t, customerEmailsNode.Error()) {
		customerEmails := customerEmailsNode.Map(func(customer Node) interface{} {
			email := customer.Get("email")
			if email.Error() != nil {
				return email.Error()
			}
			return email.String()
		})
		if assert.NoError(t, customerEmails.Error()) {
			assert.Equal(t, 2, customerEmails.Len())
		}
	}

	// Business scenario 9: Calculate total inventory value
	products = root.Get("store").Get("products") // Get fresh reference
	totalValue := 0.0
	if assert.NoError(t, products.Error()) {
		products.ForEach(func(_ interface{}, product Node) {
			priceNode := product.Get("price")
			stockNode := product.Get("stock")
			if assert.NoError(t, priceNode.Error()) && assert.NoError(t, stockNode.Error()) {
				price, _ := priceNode.RawFloat()
				stock := float64(stockNode.Int())
				totalValue += price * stock
			}
		})
	}
	// Expected: 1299.99*15 + 29.99*95 + 89.99*0 + 39.99*50
	expectedValue := 1299.99*15 + 29.99*95 + 89.99*0 + 39.99*50
	// Use tolerance for floating point comparison
	if diff := math.Abs(expectedValue - totalValue); diff > 1e-6 {
		assert.InDelta(t, expectedValue, totalValue, 1e-6, "inventory total mismatch (diff=%v)", diff)
	}

	// Business scenario 10: Error handling
	nonExistent := root.Query("/store/nonexistent/path")
	assert.Error(t, nonExistent.Error())
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
	root.RegisterFunc("active", func(n Node) Node {
		return n.Filter(func(user Node) bool {
			return user.Get("active").Bool()
		})
	})

	root.RegisterFunc("admins", func(n Node) Node {
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

	root.RegisterFunc("recentLogin", func(n Node) Node {
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
	if assert.NoError(t, activeUsers.Error()) {
		assert.Equal(t, 2, activeUsers.Len())
	}

	// Scenario 2: Get admin users
	adminUsers := root.Query("/data/users[@admins]")
	if assert.NoError(t, adminUsers.Error()) {
		assert.Equal(t, 1, adminUsers.Len())
		usernameNode := adminUsers.Index(0).Get("username")
		if assert.NoError(t, usernameNode.Error()) {
			assert.Equal(t, "admin_user", usernameNode.String())
		}
	}

	// Scenario 3: Get recently logged in users
	recentUsers := root.Query("/data/users[@recentLogin]")
	if assert.NoError(t, recentUsers.Error()) {
		assert.Equal(t, 2, recentUsers.Len()) // john_doe and admin_user
	}

	// Scenario 4: Get user full names
	fullNamesNode := root.Query("/data/users")
	if assert.NoError(t, fullNamesNode.Error()) {
		fullNames := fullNamesNode.Map(func(user Node) interface{} {
			profile := user.Get("profile")
			if profile.Error() != nil {
				return profile.Error()
			}
			firstName := profile.Get("firstName")
			if firstName.Error() != nil {
				return firstName.Error()
			}
			lastName := profile.Get("lastName")
			if lastName.Error() != nil {
				return lastName.Error()
			}
			return firstName.String() + " " + lastName.String()
		})

		if assert.NoError(t, fullNames.Error()) {
			assert.Equal(t, 3, fullNames.Len())
			names := make([]string, fullNames.Len())
			for i := 0; i < fullNames.Len(); i++ {
				nameNode := fullNames.Index(i)
				if assert.NoError(t, nameNode.Error()) {
					names[i] = nameNode.String()
				}
			}
			assert.Contains(t, names, "John Doe")
			assert.Contains(t, names, "Jane Smith")
			assert.Contains(t, names, "Admin User")
		}
	}

	// Scenario 5: Count users by theme preference
	darkThemeCount := 0
	lightThemeCount := 0
	usersNode := root.Query("/data/users")
	if assert.NoError(t, usersNode.Error()) {
		usersNode.ForEach(func(_ interface{}, user Node) {
			themeNode := user.Get("profile").Get("preferences").Get("theme")
			if assert.NoError(t, themeNode.Error()) {
				switch themeNode.String() {
				case "dark":
					darkThemeCount++
				case "light":
					lightThemeCount++
				}
			}
		})
	}
	assert.Equal(t, 2, darkThemeCount)
	assert.Equal(t, 1, lightThemeCount)

	// Scenario 6: Update user data
	users := root.Get("data").Get("users")
	if assert.NoError(t, users.Error()) {
		for i := 0; i < users.Len(); i++ {
			user := users.Index(i)
			if !assert.NoError(t, user.Error()) {
				continue
			}
			usernameNode := user.Get("username")
			if !assert.NoError(t, usernameNode.Error()) {
				continue
			}
			if usernameNode.String() == "jane_smith" {
				setNode := user.Set("active", true)
				if assert.NoError(t, setNode.Error()) {
					activeNode := user.Get("active")
					if assert.NoError(t, activeNode.Error()) {
						assert.True(t, activeNode.Bool())
					}
				}
				break
			}
		}
	}

	// Scenario 7: Add a new role to a user
	users = root.Get("data").Get("users") // Refresh reference
	if assert.NoError(t, users.Error()) {
		for i := 0; i < users.Len(); i++ {
			user := users.Index(i)
			if !assert.NoError(t, user.Error()) {
				continue
			}

			usernameNode := user.Get("username")
			if !assert.NoError(t, usernameNode.Error()) {
				continue
			}
			if usernameNode.String() == "admin_user" {
				// Get current roles and add a new one
				rolesNode := user.Get("roles")
				if !assert.NoError(t, rolesNode.Error()) {
					continue
				}

				currentRoles := rolesNode.Array()
				newRoles := make([]interface{}, len(currentRoles)+1)
				for j, role := range currentRoles {
					newRoles[j] = role.String()
				}
				newRoles[len(newRoles)-1] = "moderator"

				assert.NoError(t, user.Set("roles", newRoles).Error())

				// Refresh reference
				user = root.Get("data").Get("users").Index(i)
				if !assert.NoError(t, user.Error()) {
					continue
				}

				updatedRolesNode := user.Get("roles")
				if assert.NoError(t, updatedRolesNode.Error()) {
					assert.Equal(t, 3, updatedRolesNode.Len())
					// Verify the new role exists
					hasModerator := false
					updatedRolesNode.ForEach(func(_ interface{}, role Node) {
						if assert.NoError(t, role.Error()) && role.String() == "moderator" {
							hasModerator = true
						}
					})
					assert.True(t, hasModerator)
				}
				break
			}
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
	root.RegisterFunc("enabledFeatures", func(n Node) Node {
		return n.Filter(func(feature Node) bool {
			return feature.Get("enabled").Bool()
		})
	})

	root.RegisterFunc("fileOutputs", func(n Node) Node {
		return n.Filter(func(output Node) bool {
			return output.Get("type").String() == "file"
		})
	})

	root.RegisterFunc("securePort", func(n Node) Node {
		return n.Filter(func(server Node) bool {
			port := server.Get("port").Int()
			return port > 1024 && port < 49151 // Registered ports range
		})
	})

	// Scenario 1: Validate required configuration sections exist
	requiredSections := []string{"app", "server", "database", "logging"}
	for _, section := range requiredSections {
		sectionNode := root.Get(section)
		assert.NoError(t, sectionNode.Error(), "Missing required section: %s", section)
		assert.True(t, sectionNode.IsValid(), "Invalid section: %s", section)
	}

	// Scenario 2: Check if app is in debug mode
	debugNode := root.Query("/app/debug")
	if assert.NoError(t, debugNode.Error()) {
		assert.False(t, debugNode.Bool())
	}

	// Scenario 3: Get server configuration
	serverHost := root.Query("/server/host")
	if assert.NoError(t, serverHost.Error()) {
		assert.Equal(t, "0.0.0.0", serverHost.String())
	}
	serverPort := root.Query("/server/port")
	if assert.NoError(t, serverPort.Error()) {
		assert.Equal(t, int64(8080), serverPort.Int())
	}

	// Scenario 4: Check TLS configuration
	tlsEnabled := root.Query("/server/tls/enabled")
	if assert.NoError(t, tlsEnabled.Error()) {
		assert.True(t, tlsEnabled.Bool())
	}

	certFile := root.Query("/server/tls/certFile")
	if assert.NoError(t, certFile.Error()) {
		assert.Equal(t, "/etc/ssl/cert.pem", certFile.String())
	}

	// Scenario 5: Get database connection info
	dbHost := root.Query("/database/host")
	if assert.NoError(t, dbHost.Error()) {
		assert.Equal(t, "localhost", dbHost.String())
	}
	dbPort := root.Query("/database/port")
	if assert.NoError(t, dbPort.Error()) {
		assert.Equal(t, int64(5432), dbPort.Int())
	}
	dbName := root.Query("/database/name")
	if assert.NoError(t, dbName.Error()) {
		assert.Equal(t, "myapp", dbName.String())
	}

	// Scenario 6: Check database pool settings
	maxConn := root.Query("/database/pool/maxConnections")
	if assert.NoError(t, maxConn.Error()) {
		assert.Equal(t, int64(20), maxConn.Int())
	}
	minConn := root.Query("/database/pool/minConnections")
	if assert.NoError(t, minConn.Error()) {
		assert.Equal(t, int64(5), minConn.Int())
	}
	maxLifetime := root.Query("/database/pool/maxLifetimeMinutes")
	if assert.NoError(t, maxLifetime.Error()) {
		assert.Equal(t, int64(60), maxLifetime.Int())
	}

	// Scenario 7: Get logging configuration
	logLevel := root.Query("/logging/level")
	if assert.NoError(t, logLevel.Error()) {
		assert.Equal(t, "info", logLevel.String())
	}
	logFormat := root.Query("/logging/format")
	if assert.NoError(t, logFormat.Error()) {
		assert.Equal(t, "json", logFormat.String())
	}

	// Scenario 8: Process logging outputs
	outputsNode := root.Query("/logging/outputs")
	if assert.NoError(t, outputsNode.Error()) {
		fileOutputs := outputsNode.Filter(func(output Node) bool {
			typeNode := output.Get("type")
			return typeNode.Error() == nil && typeNode.String() == "file"
		})
		if assert.NoError(t, fileOutputs.Error()) {
			assert.Equal(t, 1, fileOutputs.Len())
		}

		stdoutOutputs := outputsNode.Filter(func(output Node) bool {
			typeNode := output.Get("type")
			return typeNode.Error() == nil && typeNode.String() == "stdout"
		})
		if assert.NoError(t, stdoutOutputs.Error()) {
			assert.Equal(t, 1, stdoutOutputs.Len())
		}
	}

	// Scenario 9: Check feature flags
	authEnabled := root.Query("/features/authentication/enabled")
	if assert.NoError(t, authEnabled.Error()) {
		assert.True(t, authEnabled.Bool())
	}
	cacheEnabled := root.Query("/features/caching/enabled")
	if assert.NoError(t, cacheEnabled.Error()) {
		assert.True(t, cacheEnabled.Bool())
	}
	monitoringEnabled := root.Query("/features/monitoring/enabled")
	if assert.NoError(t, monitoringEnabled.Error()) {
		assert.True(t, monitoringEnabled.Bool())
	}

	// Scenario 10: Get OAuth2 providers
	oauthProviders := root.Query("/features/authentication/providers")
	if assert.NoError(t, oauthProviders.Error()) {
		assert.Equal(t, 2, oauthProviders.Len())

		hasLocal := false
		hasOauth2 := false
		oauthProviders.ForEach(func(_ interface{}, provider Node) {
			if assert.NoError(t, provider.Error()) {
				switch provider.String() {
				case "local":
					hasLocal = true
				case "oauth2":
					hasOauth2 = true
				}
			}
		})
		assert.True(t, hasLocal)
		assert.True(t, hasOauth2)
	}

	// Scenario 11: Get OAuth2 configuration
	googleClientId := root.Query("/features/authentication/oauth2/google/clientId")
	if assert.NoError(t, googleClientId.Error()) {
		assert.Equal(t, "google-client-id", googleClientId.String())
	}
	githubClientId := root.Query("/features/authentication/oauth2/github/clientId")
	if assert.NoError(t, githubClientId.Error()) {
		assert.Equal(t, "github-client-id", githubClientId.String())
	}

	// Scenario 12: Update configuration
	// Change log level to debug
	logging := root.Get("logging")
	if assert.NoError(t, logging.Error()) {
		assert.NoError(t, logging.Set("level", "debug").Error())
		logLevel := root.Query("/logging/level")
		if assert.NoError(t, logLevel.Error()) {
			assert.Equal(t, "debug", logLevel.String())
		}
	}

	// Add a new OAuth2 provider
	oauth2 := root.Get("features").Get("authentication").Get("oauth2")
	if assert.NoError(t, oauth2.Error()) {
		assert.NoError(t, oauth2.Set("facebook", map[string]interface{}{
			"clientId":     "facebook-client-id",
			"clientSecret": "facebook-client-secret",
			"redirectUrl":  "https://myapp.com/auth/facebook/callback",
		}).Error())

		facebookClientId := root.Query("/features/authentication/oauth2/facebook/clientId")
		if assert.NoError(t, facebookClientId.Error()) {
			assert.Equal(t, "facebook-client-id", facebookClientId.String())
		}
	}

	// Scenario 13: Add a new logging output
	newOutput := map[string]interface{}{
		"type": "syslog",
		"host": "localhost",
		"port": 514,
	}

	loggingOutputs := root.Get("logging").Get("outputs")
	if assert.NoError(t, loggingOutputs.Error()) {
		appendResult := loggingOutputs.Append(newOutput)
		assert.NoError(t, appendResult.Error())
	}

	// Refresh the logging outputs node
	loggingOutputs = root.Get("logging").Get("outputs")
	if assert.NoError(t, loggingOutputs.Error()) {
		t.Logf("Logging outputs length: %d", loggingOutputs.Len())
		assert.Equal(t, 3, loggingOutputs.Len())

		// Find the new syslog output
		foundSyslog := false
		for i := 0; i < loggingOutputs.Len(); i++ {
			output := loggingOutputs.Index(i)
			if !assert.NoError(t, output.Error()) {
				continue
			}
			typeNode := output.Get("type")
			if !assert.NoError(t, typeNode.Error()) {
				continue
			}
			t.Logf("Output %d type: %s", i, typeNode.String())
			if typeNode.String() == "syslog" {
				foundSyslog = true
				break
			}
		}
		assert.True(t, foundSyslog)
	}
}
