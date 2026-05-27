const API = "http://localhost:8080";

export async function post(path: string) {
  await fetch(`${API}${path}`, {
    method: "POST",
  });
}

export async function getState() {
  return fetch(`${API}/state`).then((r) => r.json());
}
