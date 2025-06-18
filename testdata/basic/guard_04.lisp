(defpurefun (vanishes! x) (== 0 x))

(defcolumns (ST :i5) (A :i16) (B :i16) (C :i16))
(defconstraint c1 (:guard ST)
  (if (== 0 A)
      (vanishes! B)
      (vanishes! C)))
