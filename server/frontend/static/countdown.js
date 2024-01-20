// start(time: HTMLTimeElement, target: Date, formatDuration: (number) => string) -> void
// start starts a countdown timer that updates the text content of the
// given element with the result of calling formatDuration with the remaining
// time in milliseconds. The countdown stops when the remaining time reaches 0, at
// which point the optional done callback is called.
export function start(time, formatDuration, done = () => {}) {
  let id;
  const target = new Date(time.getAttribute("datetime")).getTime();
  const update = () => {
    const now = new Date().getTime();
    const distance = target - now;
    time.textContent = formatDuration(distance);

    if (distance <= 0) {
      clearInterval(id);
      done();
    }
  };
  id = setInterval(update, 1000);
  update();
}

// formatDurationClock formats a duration in milliseconds as a colon-separated
// string of the form "HH:MM:SS".
export function formatDurationClock(duration) {
  const h = Math.floor(duration / (1000 * 60 * 60));
  const m = Math.floor((duration % (1000 * 60 * 60)) / (1000 * 60));
  const s = Math.floor((duration % (1000 * 60)) / 1000);
  const parts = [h, m, s].map((n) => n.toString().padStart(2, "0"));
  return parts.join(":");
}

// formatDurationString formats a duration in milliseconds as a string of the
// form "Xh Ym Zs" where X, Y, and Z are integers.
export function formatDurationString(duration) {
  const h = Math.floor(duration / (1000 * 60 * 60));
  const m = Math.floor((duration % (1000 * 60 * 60)) / (1000 * 60));
  const s = Math.floor((duration % (1000 * 60)) / 1000);

  if (h > 0) {
    time.textContent = `${h}h ${m}m ${s}s`;
  } else if (m > 0) {
    time.textContent = `${m}m ${s}s`;
  } else {
    time.textContent = `${s}s`;
  }
}
