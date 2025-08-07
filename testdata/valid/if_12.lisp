(defpurefun (vanishes! x) (== 0 x))

(defcolumns (A :binary) (B :i16) (C :i16))
(defconstraint c1 ()
  (if (== 0 A)
      (vanishes! B)
      (vanishes! C)))
