import { vitePreprocess } from "@sveltejs/kit/vite";
import preprocess from "svelte-preprocess";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs#compile-time-svelte-preprocess
  // for more information about preprocessors
  preprocess: [
    vitePreprocess({}),
    /*
    preprocess({
      postcss: true,
    }),
    */
  ],
};

export default config;
