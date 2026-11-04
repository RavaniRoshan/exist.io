'use client';

import { useEffect, useState } from 'react';

export function LiveClock() {
  const [time, setTime] = useState<string | null>(null);

  useEffect(() => {
    const tick = () => {
      const now = new Date();
      setTime(
        `${now.toISOString().slice(11, 19)} utc`,
      );
    };
    tick();
    const timer = window.setInterval(tick, 1000);
    return () => window.clearInterval(timer);
  }, []);

  return (
    <span className="mono tabular-nums" suppressHydrationWarning>
      {time ?? '--:--:-- utc'}
    </span>
  );
}
