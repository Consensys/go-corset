(defcolumns (X :i32) (Y :i32))

(defconstraint c1 ()
  (if (== X 0)
      (== X Y)
      (== 1 0)))
