(defcolumns (A :i16) (B :i16))
(defconstraint c1 ()
  (if (== A 0)
      (== 0 B)))
