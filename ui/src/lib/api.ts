import type { Device } from "$lib/device";

export async function post(device: Device, path: string) {
  await fetch(`${device.baseUrl}${path}`, {
    method: "POST",
  });
}

export async function postContent(device: Device, path: string, data: object) {
  await fetch(`${device.baseUrl}${path}`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getState(device: Device) {
  return fetch(`${device.baseUrl}/state`).then((r) => r.json());
}
