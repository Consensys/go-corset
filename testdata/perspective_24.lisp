(defpurefun ((vanishes! :@loob :force) e0) e0)
;;
(defcolumns
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

;; Section 1
(defperspective p1 P ((B :binary :array [2])))
(defconstraint c1 (:perspective p1) (vanishes! (- [B 1] [B 2])))

;; Section 2
(defperspective p2 Q ((C :binary :array [2])))
(defconstraint c2 (:perspective p2) (vanishes! (* [C 1] [C 2])))
