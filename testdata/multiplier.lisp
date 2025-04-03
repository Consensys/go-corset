(defcolumns
  ;; nibbles
  (ARG1 :i4@prove :array [0:1])
  (ARG2 :i4@prove :array [0:1])
  (RES :i4@prove :array [0:3]))

(defconstraint MUL_1 ()
  (== [RES 0] (* [ARG1 0] [ARG2 0])))

(defconstraint MUL_2 ()
  (== [RES 1] (+ (* [ARG1 0] [ARG2 1]) (* [ARG1 1] [ARG2 0]))))

(defconstraint MUL_3 ()
  (== [RES 2] (* [ARG1 1] [ARG2 1])))
