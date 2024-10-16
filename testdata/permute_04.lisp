(defcolumns
  (ST :u16)
  (X :u16))
(permute (ST' Y) (+ST -X))
(vanish:last first-row (- Y 5))
