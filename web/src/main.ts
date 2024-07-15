import "@picocss/pico/css/pico.pumpkin.min.css";
import App from "./App.svelte";

const app = new App({
  target: document.getElementById("app")!,
});

export default app;
