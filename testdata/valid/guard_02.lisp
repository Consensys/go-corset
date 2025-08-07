(defpurefun (vanishes! x) (== 0 x))

(defcolumns (ST :i4) (A :i16) (B :i16))
;; STAMP == 0 || X == 1 || X == 2
(defconstraint c1 (:guard ST)
  (if (== 0 A)
      (vanishes! B)))
