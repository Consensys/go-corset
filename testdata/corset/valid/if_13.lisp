(defpurefun ((force-bin :binary :force) x) x)

(defcolumns (A :i16) (B :i16) (C :i16))
(defconstraint c1 ()
  (if (== 0 (force-bin A))
      (== 0 B)
      (== 0 C)))
