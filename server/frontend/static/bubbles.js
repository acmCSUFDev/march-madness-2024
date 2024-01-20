let bubbles;
const localStorageKey = "bubbles-enabled";

function enableBubbles() {
  if (bubbles) {
    return;
  }

  bubbles = document.createElement("div");
  bubbles.id = "bubbles-background";
  document.body.appendChild(bubbles);

  const updateHeight = () => {
    bubbles.style.setProperty("--page-height", document.body.scrollHeight + "px");
  };

  const resizeObserver = new ResizeObserver(() => updateHeight());
  resizeObserver.observe(document.body);
  window.addEventListener("resize", updateHeight);
  updateHeight();

  const bubbleParameter = {
    left: () => Math.random() * 100 - 5 + "%",
    size: () => generateVariance(100, -40, 80) + "px",
    duration: () => generateVariance(15, -10, 10) + "s",
    "wobble-scale": () => generateVariance(1, -0.5, 0.5),
    "wobble-duration": () => generateVariance(2, -1, 1) + "s",
  };

  const bubbleCount = () => bubbles.childElementCount;
  const maxBubbles = 10;
  const initialBubbles = 6;

  const spawn = () => {
    if (!bubbles) {
      return;
    }

    const bubble = document.createElement("div");
    bubble.classList.add("bubble");
    for (const [key, value] of Object.entries(bubbleParameter)) {
      bubble.style.setProperty(`--${key}`, value());
    }

    // Spawn the bubble underneath the screen wrt the scroll position.
    // Don't spawn at the bottom of the page, because it will take a while to
    // float up.
    const bubbleSize = parseFloat(bubble.style.getPropertyValue("--size"));
    const startY = window.scrollY + window.innerHeight + bubbleSize;
    console.debug("spawning bubble", {
      scrollY: window.scrollY,
      innerHeight: window.innerHeight,
      bubbleSize: bubbleSize,
    });

    bubble.style.setProperty("--startY", startY + "px");

    bubbles.appendChild(bubble);

    const duration = parseFloat(bubble.style.getPropertyValue("--duration"));
    setTimeout(() => bubble.remove(), duration * 1000);
  };

  const id = setInterval(() => {
    if (!bubbles) {
      clearInterval(id);
      return;
    }
    if (bubbleCount() < maxBubbles) {
      spawn();
    }
  }, 2000);

  const spawnDelay = () => generateVariance(0, 0, 5000);
  for (let i = 0; i < initialBubbles; i++) {
    setTimeout(spawn, spawnDelay());
  }
}

function disableBubbles() {
  if (bubbles) {
    const old = bubbles;
    bubbles = null;

    old.classList.add("disabled");
    setTimeout(() => old.remove(), 1000);
  }
}

function generateVariance(value, min, max) {
  min = value + min;
  max = value + max;
  return Math.random() * (max - min) + min;
}

export function toggleBubbles() {
  if (bubbles) {
    disableBubbles();
    window.localStorage.setItem(localStorageKey, "false");
  } else {
    enableBubbles();
    window.localStorage.setItem(localStorageKey, "true");
  }
}

export function enabled() {
  return !!bubbles;
}

const preference = window.localStorage.getItem(localStorageKey);
if (preference != null) {
  if (preference === "true") {
    enableBubbles();
  } else {
    disableBubbles();
  }
} else {
  const isReduced =
    window.matchMedia(`(prefers-reduced-motion: reduce)`) === true ||
    window.matchMedia(`(prefers-reduced-motion: reduce)`).matches === true;
  if (!isReduced) {
    enableBubbles();
  }
}
