import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/kit/vite';

/** @type {import('@sveltejs/kit').Config} */
export default {
  preprocess: vitePreprocess(),
  kit: {
		adapter: adapter({
      // Static needs a fallback page.
      fallback: 'index.html'
    }),
	},
};
