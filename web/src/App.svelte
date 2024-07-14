<script lang="ts">
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
  let lastUpdated: string | undefined = undefined;

  const containerUrl = (container: Container): string => {
    const url = new URL(window.origin);
    url.hostname = container.name + "." + url.hostname;
    return url.toString();
  }

  const refresh = async () => {
    const json = await fetch("/api/containers").then(r => r.json());
    containers = Container.array().parse(json);
    lastUpdated = new Date().toLocaleString();
  }

  setInterval(async () => {
    await refresh();
  }, 60 * 1000)

  // initial fetch
  refresh();
</script>

<main class="wrapper">
  <nav class="navigation">
    <section class="container">
      <span class="title">Karja</span>
    </section>
  </nav>
  <section class="container">
    <h3>Running containers</h3>
    <p>Last updated: {lastUpdated ?? "N/A"}</p>
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
  </section>
</main>

<style lang="sass">
  .navigation
    max-width: 100%
    background: #f4f4f4
    margin-bottom: 1rem
    .container
      display: flex
      align-items: center
      height: 5.2rem
      .title
        font-size: 1.2em

  .container
    margin: 0 auto
    width: 80rem
</style>
