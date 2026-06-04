Feature: Function data and scheduling
  Functions can keep values in a key-value store and run on a schedule. Both are
  configured from the dashboard.

  Background:
    Given I am signed in
    And I create a function named "stateful"
    And I open the function "stateful"

  Scenario: Storing a value in the key-value store
    When I store the value "ready" under the key "status"
    Then the stored value "ready" should be kept under the key "status"

  Scenario: Storing a value in the global store
    When I store the value "shared-secret" under the global key "token"
    Then the global value "shared-secret" should be kept under the key "token"

  Scenario: A stored value is readable from the function
    Given I store the value "blue" under the key "color"
    And I change the function's code to:
      """
      function handler(ctx, event)
        return { statusCode = 200, body = kv.get("color") or "none" }
      end
      """
    When I call the function
    Then the response should be "blue"

  Scenario: Putting a function on a schedule
    When I schedule the function to run every 5 minutes
    Then I should see when the function will next run
