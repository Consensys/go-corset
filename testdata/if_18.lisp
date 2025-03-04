(defcolumns (X :@loob) (Y :@loob))

(defconstraint c1 ()
  (if X
      (if (shift X -1)
          Y)))
