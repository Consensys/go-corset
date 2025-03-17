(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (if (== X 0)
      (if (== 0 (shift X -1))
          (== 0 Y))))
