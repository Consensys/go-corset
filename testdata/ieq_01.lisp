(defcolumns (ARG_1 :i16) (ARG_2 :i16) (RES :binary@prove))

(defconstraint c1 ()
  (if (< ARG_1 ARG_2)
      (== RES 1)
      (== RES 0)))
