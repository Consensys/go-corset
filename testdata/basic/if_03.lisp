(defpurefun (vanishes! x) (== 0 x))

(defcolumns (A :i16) (B :i16))
(defconstraint c1 ()
  (if (== 0 A)
      (vanishes! 0)
      (vanishes! B)))

(defconstraint c2 ()
  (if (!= 0 A)
      (vanishes! B)
      (vanishes! 0)))
