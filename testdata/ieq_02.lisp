(defcolumns (X :i16) (Y :i16))
(defconst FORK 3)

(defconstaint c1 ()
  (if (< FORK 5)
      (== X (+ Y Y))))

(defconstaint c2 ()
  (if (> FORK 5)
      (== X Y)))
