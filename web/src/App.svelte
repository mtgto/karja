<script lang="ts">
  import { z } from "zod";

  const Container = z.object({
    id: z.string().min(1),
    name: z.string().min(1),
    status: z.string().min(1),
    healthy: z.boolean(),
    connectable: z.boolean(),
  });

  type Container = z.infer<typeof Container>;
  let containers: Container[] = [];
  let lastUpdated: string | undefined = undefined;
  // next schedule to fetch
  let scheduled: Date = new Date();
  let remainingSecond: number = 0;

  const containerUrl = (container: Container): string => {
    const url = new URL(window.origin);
    url.hostname = container.name + "." + url.hostname;
    return url.toString();
  }

  const refresh = async () => {
    const json = await fetch("/api/containers").then(r => r.json());
    containers = Container.array().parse(json).sort((a, b) => {
      // sort by connectable
      if (a.connectable === b.connectable) {
        return 0;
      } else if (a.connectable) {
        return -1;
      } else {
        return 1;
      }
    });
    lastUpdated = new Date().toLocaleString();
    scheduled = new Date(Date.now() + 60000);
    remainingSecond = 60;
  }

  setInterval(async () => {
    const now = Date.now();
    if (scheduled.getTime() < now) {
      try {
        await refresh();
      } catch (e) {
        console.error(e)
        scheduled = new Date(Date.now() + 60000);
        remainingSecond = 60;
      }
    } else {
      remainingSecond = Math.floor((scheduled.getTime() - now) / 1000)
    }
  }, 1000)
</script>

<main class="wrapper">
  <header class="navigation">
    <nav class="container">
      <ul>
        <li><strong>Karja</strong></li>
      </ul>
      <ul>
        <li><a href="https://github.com/mtgto/karja">GitHub</a></li>
      </ul>
    </nav>
  </header>
  <section class="container">
    <h3>Running containers</h3>
    <p>Last updated: {lastUpdated ?? "N/A"}</p>
    <table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Status</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        {#each containers as container}
          <tr>
            <td>{container.name}</td>
            <td>{container.status}</td>
            <td>
              {#if container.connectable}
              <a target="_blank" href="{containerUrl(container)}">Open</a>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
    <section class="schedule">
      <p>Next update will begin {remainingSecond} seconds later.</p>
      <button on:click={refresh}>Update Now</button>
    </section>
  </section>
</main>

<style lang="scss">
  :root {
    --pico-form-element-spacing-vertical: 0.4rem;
  }
  header.navigation {
    margin-bottom: 1rem;
  }
  .schedule {
    margin-top: 2rem;
  }
</style>
