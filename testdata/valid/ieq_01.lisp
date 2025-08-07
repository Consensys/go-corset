(defcolumns (ARG_1 :i16) (ARG_2 :i16) (RES :binary@prove))

(defconstraint lt ()
  (if (< ARG_1 ARG_2)
      (== RES 1)
      (== RES 0)))

(defconstraint lteq ()
  (if (<= ARG_2 ARG_1)
      (== RES 0)
      (== RES 1)))

(defconstraint gt ()
  (if (> ARG_2 ARG_1)
      (== RES 1)
      (== RES 0)))

(defconstraint gteq ()
  (if (>= ARG_1 ARG_2)
      (== RES 0)
      (== RES 1)))
