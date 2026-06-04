Feature: Connected clients
  Operators can review the API clients (such as the CLI) connected to the server.

  Background:
    Given I am signed in

  Scenario: Viewing the connected clients page
    When I open the connected clients page
    Then the connected clients page should be shown
