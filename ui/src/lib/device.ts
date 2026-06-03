export type Device = {
  id: string;
  label: string;
  baseUrl: string;
};

let currentDevice: Device | null = null;

async function probe(baseUrl: string): Promise<Device | null> {
  try {
    const response = await fetch(`${baseUrl}/identity`);

    if (!response.ok) {
      return null;
    }

    const identity = await response.json();

    return {
      id: identity.id,
      label: identity.label,
      baseUrl,
    };
  } catch {
    return null;
  }
}


