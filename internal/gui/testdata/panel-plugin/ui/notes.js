// A panel in each provider's editor: starling.context.provider says which
// provider it is in, and the note is kept in the plugin's own options.
(async () => {
  await starling.ready;
  const id = starling.context.provider.id;
  let opts = await starling.options.get();
  const box = document.getElementById("note");
  box.value = opts.notes?.[id] || "";
  let timer = 0;
  // options are set once a second at most: the last keystroke's, a
  // little after it
  box.oninput = () => {
    clearTimeout(timer);
    timer = setTimeout(async () => {
      opts = await starling.options.set({ ...opts, notes: { ...opts.notes, [id]: box.value } });
      document.getElementById("saved").textContent = "Saved";
    }, 1200);
  };
})();
