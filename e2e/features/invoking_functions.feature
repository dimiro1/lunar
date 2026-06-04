Feature: Invoking functions
  A deployed function is reachable over HTTP. These scenarios create a function
  in the dashboard and then call it the way a real client would.

  Background:
    Given I am signed in

  Scenario: A brand-new function responds out of the box
    Given I create a function named "hello"
    And I open the function "hello"
    When I call the function
    Then the function should respond successfully
    And the response should contain "Hello from Lua!"

  Scenario: A Starlark function responds too
    Given I create a Starlark function named "hello-star"
    And I open the function "hello-star"
    When I call the function
    Then the function should respond successfully
    And the response should contain "Hello from Starlark!"

  Scenario: A function built from the REST API template responds
    Given I create a function named "rest" from the "api" template
    And I open the function "rest"
    When I call the function
    Then the function should respond successfully

  Scenario: A function can be called with different HTTP methods
    Given I create a function named "methods"
    And I open the function "methods"
    And I change the function's code to:
      """
      function handler(ctx, event)
        return { statusCode = 200, body = event.method }
      end
      """
    When I call the function with the POST method
    Then the function should respond successfully
    And the response should be "POST"

  Scenario: A disabled function cannot be called
    Given I create a function named "switched-off"
    And I open the function "switched-off"
    When I disable the function
    And I call the function
    Then the call should be refused because the function is disabled

  Scenario: A deleted function cannot be called
    Given I create a function named "gone"
    And I open the function "gone"
    When I delete the function
    And I call the function
    Then the function should no longer be reachable

  Scenario: Environment variables are available to the function
    Given I create a function named "greeter-env"
    And I open the function "greeter-env"
    And I change the function's code to:
      """
      function handler(ctx, event)
        return { statusCode = 200, body = env.get("GREETING") or "unset" }
      end
      """
    And I give the function an environment variable "GREETING" set to "hola"
    When I call the function
    Then the response should be "hola"
