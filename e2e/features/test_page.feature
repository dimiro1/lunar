Feature: Trying a function from the dashboard
  The dashboard's Test page lets operators send a request to a function and see
  the response without leaving the browser.

  Background:
    Given I am signed in
    And I create a function named "playground"
    And I open the function "playground"

  Scenario: Running a function from the Test page
    When I try the function from the Test page
    Then the Test page should show a successful response
    And the Test page response should contain "Hello from Lua!"
