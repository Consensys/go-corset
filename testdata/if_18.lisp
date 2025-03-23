(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (if (== X 0)
      (if (== (shift X -1) 0)
          (== 0 Y))))
