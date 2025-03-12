(defcolumns (X :i16@loob) (Y :i16@loob))

(defconstraint c1 ()
  (if X
      (if (shift X -1)
          Y)))
