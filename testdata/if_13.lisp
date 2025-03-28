(defpurefun (vanishes! x) (== 0 x))
(defpurefun ((force-bin :binary :force) x) x)

(defcolumns (A :i16) (B :i16) (C :i16))
(defconstraint c1 ()
  (if (vanishes! (force-bin A))
      (vanishes! B)
      (vanishes! C)))
