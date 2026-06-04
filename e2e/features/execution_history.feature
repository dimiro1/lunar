Feature: Execution history
  Every call to a function is recorded so operators can audit what happened.

  Background:
    Given I am signed in
    And I create a function named "audited"
    And I open the function "audited"

  Scenario: A function that has never run has no history
    When I view the function's execution history
    Then the execution history should be empty

  Scenario: Calling a function records an execution
    When I call the function
    And I view the function's execution history
    Then the function should have 1 recorded execution
    And the latest execution should be successful

  Scenario: Each call is recorded separately
    When I call the function
    And I call the function
    And I call the function
    Then the function should have 3 recorded executions

  Scenario: Inspecting a recorded execution
    When I call the function
    And I open the most recent execution
    Then the execution should be shown as successful
