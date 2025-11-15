export const API = {
  // Functions
  functions: {
    list: (limit = 20, offset = 0) =>
      m.request({
        method: "GET",
        url: `/api/functions?limit=${limit}&offset=${offset}`,
      }),
    get: (id) => m.request({ method: "GET", url: `/api/functions/${id}` }),
    create: (data) =>
      m.request({ method: "POST", url: "/api/functions", body: data }),
    update: (id, data) =>
      m.request({ method: "PUT", url: `/api/functions/${id}`, body: data }),
    delete: (id) =>
      m.request({ method: "DELETE", url: `/api/functions/${id}` }),
    updateEnv: (id, env_vars) =>
      m.request({
        method: "PUT",
        url: `/api/functions/${id}/env`,
        body: { env_vars },
      }),
  },

  // Versions
  versions: {
    list: (functionId, limit = 20, offset = 0) =>
      m.request({
        method: "GET",
        url: `/api/functions/${functionId}/versions?limit=${limit}&offset=${offset}`,
      }),
    get: (functionId, version) =>
      m.request({
        method: "GET",
        url: `/api/functions/${functionId}/versions/${version}`,
      }),
    activate: (functionId, version) =>
      m.request({
        method: "POST",
        url: `/api/functions/${functionId}/versions/${version}/activate`,
      }),
    diff: (functionId, v1, v2) =>
      m.request({
        method: "GET",
        url: `/api/functions/${functionId}/diff/${v1}/${v2}`,
      }),
  },

  // Executions
  executions: {
    list: (functionId, limit = 20, offset = 0) =>
      m.request({
        method: "GET",
        url: `/api/functions/${functionId}/executions?limit=${limit}&offset=${offset}`,
      }),
    get: (executionId) =>
      m.request({ method: "GET", url: `/api/executions/${executionId}` }),
    getLogs: (executionId, limit = 20, offset = 0) =>
      m.request({
        method: "GET",
        url: `/api/executions/${executionId}/logs?limit=${limit}&offset=${offset}`,
      }),
  },

  // Execute function
  execute: (functionId, request) => {
    // Handle query as either string or object
    let queryString = "";
    if (request.query) {
      if (typeof request.query === "string") {
        queryString = request.query.startsWith("?")
          ? request.query
          : "?" + request.query;
      } else {
        const params = new URLSearchParams(request.query);
        queryString = params.toString() ? "?" + params : "";
      }
    }
    const url = `/fn/${functionId}${queryString}`;

    return m.request({
      method: request.method || "GET",
      url: url,
      body: request.body,
      headers: request.headers,
      extract: (xhr) => ({
        status: xhr.status,
        body: xhr.responseText,
        headers: {
          "X-Function-Id": xhr.getResponseHeader("X-Function-Id"),
          "X-Function-Version-Id": xhr.getResponseHeader(
            "X-Function-Version-Id",
          ),
          "X-Execution-Id": xhr.getResponseHeader("X-Execution-Id"),
          "X-Execution-Duration-Ms": xhr.getResponseHeader(
            "X-Execution-Duration-Ms",
          ),
        },
      }),
    });
  },
};
