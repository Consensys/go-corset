(column ST :u16)
(column X :u16)
(permute (ST' Y) (+ST -X))
(vanish:last first-row (- Y 5))
