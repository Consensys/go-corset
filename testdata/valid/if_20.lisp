(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (if (== X 0)
      (== X Y)
      (== 1 0)))
