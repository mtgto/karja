<script lang="ts">
  import Counter from "./lib/Counter.svelte";
  import { z } from "zod";

  const Container = z.object({
    id: z.string().min(1),
    name: z.string().min(1),
    status: z.string().min(1),
    public_port: z.number().positive().lte(65535),
    private_port: z.number().positive().lte(65535),
    healthy: z.boolean(),
  });

  type Container = z.infer<typeof Container>;
  let containers: Container[] = [];

  const containerUrl = (container: Container): string => {
    const url = new URL(window.origin)
    url.hostname = container.name + "." + url.hostname
    return url.toString()
  }

  const refresh = async () => {
    const json = await fetch("/api/containers").then(r => r.json())
    containers = Container.array().parse(json)
  }

  setInterval(async () => {
    await refresh()
  }, 60 * 1000)

  // initial fetch
  refresh()
</script>

<main>
  <section class="container">
    <h2>Karja</h2>

    <h3>Running containers</h3>
    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Port</th>
          <th>Status</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        {#each containers as container}
          <tr>
            <td>{container.name}</td>
            <td>{container.public_port}:{container.private_port}</td>
            <td>{container.status}</td>
            <td><a target="_blank" href="{containerUrl(container)}">Open</a></td>
          </tr>
        {/each}
      </tbody>
    </table>

    <div class="card">
      <Counter />
    </div>

    <p>
      Check out <a href="https://github.com/sveltejs/kit#readme" target="_blank" rel="noreferrer">SvelteKit</a>, the official Svelte app framework powered by Vite!
    </p>

    <p class="read-the-docs">
      Click on the Vite and Svelte logos to learn more
    </p>
  </section>
</main>

<style>
  .container {
    margin: 0 auto;
    width: 80rem;
  }
</style>
