(defpurefun (vanishes! x) (== 0 x))

(defcolumns (A :i16) (B :i16))
(defconstraint c1 ()
  (if (== A 0)
      (vanishes! B)))
