Feature: Signing in to the dashboard
  The dashboard is protected by an API key. Operators sign in to manage their
  functions and can sign out when they are done.

  Scenario: Signing in with a valid key
    When I sign in with a valid API key
    Then I should reach the dashboard

  Scenario: Signing in with an invalid key
    When I sign in with an invalid API key
    Then I should be told the key is invalid

  Scenario: Signing out
    Given I am signed in
    When I sign out
    Then I should be signed out
