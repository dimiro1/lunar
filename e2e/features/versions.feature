Feature: Working with versions
  Every code change is saved as a new version. Operators can review the history,
  roll back to an earlier version, and compare two versions side by side.

  Background:
    Given I am signed in
    And I create a function named "evolving"
    And I open the function "evolving"

  Scenario: Editing the code creates a new version
    When I change the function's code to:
      """
      function handler(ctx, event)
        return { statusCode = 200, body = "second cut" }
      end
      """
    Then the function should have 2 versions
    And version 2 should be the active version

  Scenario: Rolling back to an earlier version
    Given I change the function's code to:
      """
      function handler(ctx, event)
        return { statusCode = 200, body = "new behaviour" }
      end
      """
    When I roll back to version 1
    Then version 1 should be the active version
    And I call the function
    And the response should contain "Hello from Lua!"

  Scenario: Comparing two versions
    Given I change the function's code to:
      """
      function handler(ctx, event)
        return { statusCode = 200, body = "the changed line" }
      end
      """
    When I compare version 1 with version 2
    Then I should see the differences between the two versions
