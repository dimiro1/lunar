Feature: Managing functions
  Operators create, find, rename, and delete functions from the dashboard.

  Background:
    Given I am signed in

  Scenario: The functions list starts empty
    Given I have not created any functions yet
    When I view my functions
    Then my functions list should be empty

  Scenario: Creating a function
    When I create a function named "greeter"
    Then I should land back on my functions list
    And I should see "greeter" in my functions list

  Scenario: The list shows which language each function uses
    Given I create a function named "lua-one"
    And I create a Starlark function named "star-one"
    When I view my functions
    Then "Lua" should be shown as a supported language
    And "Starlark" should be shown as a supported language

  Scenario: A new function needs a name
    When I try to create a function without giving it a name
    Then I should be told the function needs a name

  Scenario: Renaming a function
    Given I create a function named "before-rename"
    And I open the function "before-rename"
    When I rename the function to "after-rename"
    And I view my functions
    Then I should see "after-rename" in my functions list

  Scenario: Deleting a function
    Given I create a function named "throwaway"
    And I open the function "throwaway"
    When I delete the function
    Then I should no longer see "throwaway" in my functions list
