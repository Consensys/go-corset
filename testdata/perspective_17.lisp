(defpurefun ((vanishes! :@loob :force) e0) e0)
;;
(defcolumns
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

;; Section 1
(defperspective p1 P ((A :binary) (B :binary) (E :byte)))
(defconstraint c1 (:perspective p1) (vanishes! (- A B)))

;; Section 2
(defperspective p2 Q ((C :binary) (D :byte)))
(defconstraint c2 (:perspective p2) (vanishes! (* C D)))
