(defpurefun (vanishes! x) (== 0 x))

(defcolumns (ST :i3) (A :i16) (B :i16))
(defconstraint c1 (:guard ST)
  (if (== 0 A)
      (vanishes! 0)
      (vanishes! B)))
