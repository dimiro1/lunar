/**
 * @fileoverview Template card components for function creation wizard.
 */

import { icons } from "../icons.js";

/**
 * @typedef {Object} FunctionTemplate
 * @property {string} id - Unique template identifier
 * @property {string} name - Display name
 * @property {string} description - Template description
 * @property {string} icon - Icon name from icons module
 * @property {string} code - Template Lua code
 */

/**
 * Template card component for function creation.
 * Displays a selectable card with icon, name, and description.
 * @type {Object}
 */
export const TemplateCard = {
  /**
   * Renders the template card component.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} vnode.attrs.name - Template name
   * @param {string} [vnode.attrs.description] - Template description
   * @param {string} [vnode.attrs.icon='code'] - Icon name
   * @param {boolean} [vnode.attrs.selected=false] - Whether card is selected
   * @param {function} [vnode.attrs.onclick] - Click handler
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const {
      name,
      description,
      icon = "code",
      selected = false,
      onclick,
      class: className = "",
    } = vnode.attrs;

    const classes = [
      "template-card",
      selected && "template-card--selected",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    return m(
      "div",
      {
        class: classes,
        onclick,
        role: "button",
        tabindex: 0,
        "aria-pressed": selected,
        onkeydown: (e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onclick && onclick(e);
          }
        },
      },
      [
        m(".template-card__header", [
          m(".template-card__icon", m.trust(icons[icon]())),
          m(".template-card__name", name),
        ]),
        description && m("p.template-card__description", description),
      ],
    );
  },
};

/**
 * Template cards grid container.
 * Provides CSS grid layout for template cards.
 * @type {Object}
 */
export const TemplateCards = {
  /**
   * Renders the template cards container.
   * @param {Object} vnode - Mithril vnode
   * @param {Object} vnode.attrs - Component attributes
   * @param {string} [vnode.attrs.class] - Additional CSS classes
   * @param {*} vnode.children - Child elements to render (TemplateCard)
   * @returns {Object} Mithril vnode
   */
  view(vnode) {
    const { class: className = "" } = vnode.attrs;

    return m(
      ".template-cards",
      {
        class: className,
      },
      vnode.children,
    );
  },
};

/**
 * Pre-defined templates for function creation.
 * Each template includes sample Lua code for common use cases.
 * @type {FunctionTemplate[]}
 */
export const FunctionTemplates = [
  {
    id: "http",
    name: "HTTP Template",
    description: "Handle HTTP requests with custom logic",
    icon: "globe",
    code: `-- HTTP Handler
function handler(ctx, event)
    local method = event.method
    local path = event.path

    log.info("Received " .. method .. " request to " .. path)

    return {
        statusCode = 200,
        headers = { ["Content-Type"] = "application/json" },
        body = json.encode({
            message = "Hello from Lua!",
            method = method,
            path = path
        })
    }
end`,
  },
  {
    id: "api",
    name: "REST API",
    description: "Build RESTful API endpoints",
    icon: "server",
    code: `-- REST API Endpoint
function handler(ctx, event)
    local method = event.method

    if method == "GET" then
        return {
            statusCode = 200,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({
                items = {},
                total = 0
            })
        }
    elseif method == "POST" then
        local data = json.decode(event.body)
        return {
            statusCode = 201,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({
                id = crypto.uuid(),
                created = true
            })
        }
    else
        return {
            statusCode = 405,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({
                error = "Method not allowed"
            })
        }
    end
end`,
  },
  {
    id: "ai-chat",
    name: "AI Chatbot",
    description: "Chat completion with OpenAI or Anthropic",
    icon: "sparkles",
    code: `-- AI Chatbot
-- Set OPENAI_API_KEY or ANTHROPIC_API_KEY in environment variables
function handler(ctx, event)
    -- Parse request body
    local data, err = json.decode(event.body)
    if err then
        return {
            statusCode = 400,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({ error = "Invalid JSON" })
        }
    end

    local message = data.message or "Hello!"

    -- Call AI provider
    local response, err = ai.chat({
        provider = "openai",  -- or "anthropic"
        model = "gpt-4o-mini",  -- or "claude-3-haiku-20240307"
        messages = {
            { role = "system", content = "You are a helpful assistant." },
            { role = "user", content = message }
        },
        max_tokens = 500
    })

    if err then
        log.error("AI error: " .. err)
        return {
            statusCode = 500,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({ error = err })
        }
    end

    log.info("AI responded with " .. response.usage.output_tokens .. " tokens")

    return {
        statusCode = 200,
        headers = { ["Content-Type"] = "application/json" },
        body = json.encode({
            reply = response.content,
            model = response.model,
            tokens = response.usage.input_tokens + response.usage.output_tokens
        })
    }
end`,
  },
  {
    id: "email",
    name: "Send Email",
    description: "Send emails via Resend API",
    icon: "mail",
    code: `-- Send Email
-- Set RESEND_API_KEY in environment variables
function handler(ctx, event)
    -- Parse request body
    local data, err = json.decode(event.body)
    if err then
        return {
            statusCode = 400,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({ error = "Invalid JSON" })
        }
    end

    -- Validate required fields
    if not data.to or not data.subject then
        return {
            statusCode = 400,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({ error = "Missing required fields: to, subject" })
        }
    end

    -- Send email via Resend
    local result, err = email.send({
        from = "noreply@yourdomain.com",  -- Update with your verified domain
        to = data.to,
        subject = data.subject,
        html = data.html or "<p>" .. (data.text or "Hello!") .. "</p>",
        text = data.text
    })

    if err then
        log.error("Email error: " .. err)
        return {
            statusCode = 500,
            headers = { ["Content-Type"] = "application/json" },
            body = json.encode({ error = err })
        }
    end

    log.info("Email sent: " .. result.id)

    return {
        statusCode = 200,
        headers = { ["Content-Type"] = "application/json" },
        body = json.encode({
            success = true,
            email_id = result.id
        })
    }
end`,
  },
  {
    id: "blank",
    name: "Blank",
    description: "Start with empty template",
    icon: "document",
    code: `-- Your function code here
function handler(ctx, event)
    return {
        statusCode = 200,
        headers = { ["Content-Type"] = "text/plain" },
        body = "Hello, World!"
    }
end`,
  },
];
