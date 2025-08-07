(defcolumns (X :i16) (Y :i16))

;; (defconstraint test1 ()
;;   (== X
;;      (if (== 0 Y) 0)))

(defconstraint test2 ()
  (== X
     (if (== 0 Y) 0 16)))
